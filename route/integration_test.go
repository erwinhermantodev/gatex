package route_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/config"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	pb "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/proto/auth"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route"
)

const (
	adminToken = "test-admin-token"
	jwtSecret  = "test-jwt-secret"
	apiKey     = "test-api-key"
)

var gateway *echo.Echo

func TestMain(m *testing.M) {
	dir, _ := os.MkdirTemp("", "gatex-it")
	defer os.RemoveAll(dir)

	os.Setenv("ADMIN_API_TOKEN", adminToken)
	os.Setenv("JWT_SECRET", jwtSecret)
	os.Setenv("API_KEYS", apiKey+",other-key")
	os.Setenv("ALLOW_LOOPBACK_UPSTREAMS", "true") // httptest/grpc servers listen on 127.0.0.1

	config.Reset() // package init already loaded config before the env above was set

	d, err := gorm.Open(sqlite.Open(filepath.Join(dir, "it.db")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err := database.UseDB(d); err != nil {
		panic(err)
	}
	gateway = route.Init()
	os.Exit(m.Run())
}

// do sends a request through the in-process gateway.
func do(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = fmt.Sprintf("10.0.%d.%d:1234", time.Now().Nanosecond()%250, time.Now().Nanosecond()/1000%250) // dodge the per-IP rate limit
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	gateway.ServeHTTP(rec, req)
	return rec
}

func admin(method, path string, body interface{}) *httptest.ResponseRecorder {
	return do(method, path, body, map[string]string{"Authorization": "Bearer " + adminToken})
}

func idOf(t *testing.T, rec *httptest.ResponseRecorder) uint {
	t.Helper()
	var out struct{ ID uint }
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.ID == 0 {
		t.Fatalf("no ID in response (%d): %s", rec.Code, rec.Body.String())
	}
	return out.ID
}

func mustCode(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, want, rec.Body.String())
	}
}

func TestAdminRequiresToken(t *testing.T) {
	mustCode(t, do(http.MethodGet, "/admin/services", nil, nil), http.StatusUnauthorized)
	mustCode(t, do(http.MethodGet, "/admin/services", nil, map[string]string{"Authorization": "Bearer wrong"}), http.StatusUnauthorized)
	mustCode(t, admin(http.MethodGet, "/admin/services", nil), http.StatusOK)
}

func newRESTService(t *testing.T, name string) (id uint, upstream *httptest.Server) {
	t.Helper()
	upstream = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "upstream:%s:%s", r.Method, r.URL.Path)
	}))
	t.Cleanup(upstream.Close)
	rec := admin(http.MethodPost, "/admin/services", map[string]interface{}{"Name": name, "BaseURL": upstream.URL, "Protocol": "rest"})
	mustCode(t, rec, http.StatusCreated)
	return idOf(t, rec), upstream
}

func TestRouteChangesApplyWithoutRestart(t *testing.T) {
	svcID, _ := newRESTService(t, "reload-svc")

	mustCode(t, do(http.MethodGet, "/api/hello", nil, nil), http.StatusNotFound)

	rec := admin(http.MethodPost, "/admin/routes", map[string]interface{}{
		"Path": "/api/hello", "Method": "GET", "ServiceID": svcID, "EndpointFilter": "hello", "Tag": "t",
	})
	mustCode(t, rec, http.StatusCreated)
	routeID := idOf(t, rec)

	got := do(http.MethodGet, "/api/hello", nil, nil)
	mustCode(t, got, http.StatusOK)
	if got.Body.String() != "upstream:GET:/api/hello" {
		t.Fatalf("unexpected body %q", got.Body.String())
	}
	mustCode(t, do(http.MethodPost, "/api/hello", nil, nil), http.StatusMethodNotAllowed)

	mustCode(t, admin(http.MethodDelete, fmt.Sprintf("/admin/routes/%d", routeID), nil), http.StatusNoContent)
	mustCode(t, do(http.MethodGet, "/api/hello", nil, nil), http.StatusNotFound)
}

func TestPathParamsAreMatched(t *testing.T) {
	svcID, _ := newRESTService(t, "param-svc")
	mustCode(t, admin(http.MethodPost, "/admin/routes", map[string]interface{}{
		"Path": "/api/users/:id", "Method": "GET", "ServiceID": svcID, "EndpointFilter": "user", "Tag": "t",
	}), http.StatusCreated)
	got := do(http.MethodGet, "/api/users/42", nil, nil)
	mustCode(t, got, http.StatusOK)
	if got.Body.String() != "upstream:GET:/api/users/42" {
		t.Fatalf("unexpected body %q", got.Body.String())
	}
}

func TestAdminValidation(t *testing.T) {
	svcID, _ := newRESTService(t, "valid-svc")
	cases := []struct {
		name, path string
		body       map[string]interface{}
	}{
		{"metadata SSRF", "/admin/services", map[string]interface{}{"Name": "evil", "BaseURL": "http://169.254.169.254/latest", "Protocol": "rest"}},
		{"non-http scheme", "/admin/services", map[string]interface{}{"Name": "evil2", "BaseURL": "file:///etc/passwd", "Protocol": "rest"}},
		{"bad protocol", "/admin/services", map[string]interface{}{"Name": "evil3", "BaseURL": "http://10.0.0.1", "Protocol": "ftp"}},
		{"grpc addr without port", "/admin/services", map[string]interface{}{"Name": "evil4", "GRPCAddr": "10.0.0.1", "Protocol": "grpc"}},
		{"reserved path", "/admin/routes", map[string]interface{}{"Path": "/admin/steal", "Method": "GET", "ServiceID": svcID}},
		{"bad method", "/admin/routes", map[string]interface{}{"Path": "/x", "Method": "TRACE", "ServiceID": svcID}},
		{"unknown service", "/admin/routes", map[string]interface{}{"Path": "/x", "Method": "GET", "ServiceID": 99999}},
		{"unknown middleware", "/admin/routes", map[string]interface{}{"Path": "/x", "Method": "GET", "ServiceID": svcID, "Middleware": `["jwtt"]`}},
		{"wildcard not last", "/admin/routes", map[string]interface{}{"Path": "/x/*/y", "Method": "GET", "ServiceID": svcID}},
		{"proto mapping on rest service", "/admin/proto-mappings", map[string]interface{}{"ServiceID": svcID, "RPCMethod": "A", "ServiceName": "B", "ProtoPackage": "c"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { mustCode(t, admin(http.MethodPost, c.path, c.body), http.StatusBadRequest) })
	}
}

func TestRouteAuthMiddlewares(t *testing.T) {
	svcID, _ := newRESTService(t, "auth-svc")
	mustCode(t, admin(http.MethodPost, "/admin/routes", map[string]interface{}{
		"Path": "/secure/key", "Method": "GET", "ServiceID": svcID, "EndpointFilter": "k", "Middleware": `["api-key"]`,
	}), http.StatusCreated)
	mustCode(t, admin(http.MethodPost, "/admin/routes", map[string]interface{}{
		"Path": "/secure/jwt", "Method": "GET", "ServiceID": svcID, "EndpointFilter": "j", "Middleware": `["jwt"]`,
	}), http.StatusCreated)

	mustCode(t, do(http.MethodGet, "/secure/key", nil, nil), http.StatusUnauthorized)
	mustCode(t, do(http.MethodGet, "/secure/key", nil, map[string]string{"x-api-key": "nope"}), http.StatusUnauthorized)
	mustCode(t, do(http.MethodGet, "/secure/key", nil, map[string]string{"x-api-key": apiKey}), http.StatusOK)

	bearer := func(tok string) map[string]string { return map[string]string{"Authorization": "Bearer " + tok} }
	mustCode(t, do(http.MethodGet, "/secure/jwt", nil, nil), http.StatusUnauthorized)
	mustCode(t, do(http.MethodGet, "/secure/jwt", nil, bearer(makeJWT("HS256", jwtSecret, time.Now().Add(time.Hour)))), http.StatusOK)
	mustCode(t, do(http.MethodGet, "/secure/jwt", nil, bearer(makeJWT("HS256", jwtSecret, time.Now().Add(-time.Hour)))), http.StatusUnauthorized)
	mustCode(t, do(http.MethodGet, "/secure/jwt", nil, bearer(makeJWT("HS256", "wrong-secret", time.Now().Add(time.Hour)))), http.StatusUnauthorized)
	mustCode(t, do(http.MethodGet, "/secure/jwt", nil, bearer(makeJWT("none", "", time.Now().Add(time.Hour)))), http.StatusUnauthorized)
}

func makeJWT(alg, secret string, exp time.Time) string {
	enc := func(v interface{}) string { b, _ := json.Marshal(v); return base64.RawURLEncoding.EncodeToString(b) }
	input := enc(map[string]string{"alg": alg, "typ": "JWT"}) + "." + enc(map[string]interface{}{"sub": "u1", "exp": exp.Unix()})
	sig := ""
	if alg == "HS256" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(input))
		sig = base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	}
	return input + "." + sig
}

// fakeAuth records the metadata it receives and answers two RPCs.
type fakeAuth struct {
	pb.UnimplementedAuthServiceServer
	mu       sync.Mutex
	gotMeta  metadata.MD
	gotPhone string
}

func (f *fakeAuth) CheckPhone(ctx context.Context, r *pb.CheckPhoneRequest) (*pb.CheckPhoneResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	f.mu.Lock()
	f.gotMeta, f.gotPhone = md, r.PhoneNumber
	f.mu.Unlock()
	return &pb.CheckPhoneResponse{Success: true, Message: "checked", Code: "OK"}, nil
}

func (f *fakeAuth) SendOTP(ctx context.Context, r *pb.SendOTPRequest) (*pb.StandardResponse, error) {
	return &pb.StandardResponse{Success: true, Message: "otp-sent"}, nil
}

func TestGRPCTranscodingPerRouteMapping(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	fake := &fakeAuth{}
	srv := grpc.NewServer()
	pb.RegisterAuthServiceServer(srv, fake)
	reflection.Register(srv)
	go srv.Serve(lis)
	t.Cleanup(srv.Stop)

	rec := admin(http.MethodPost, "/admin/services", map[string]interface{}{"Name": "grpc-svc", "GRPCAddr": lis.Addr().String(), "Protocol": "grpc"})
	mustCode(t, rec, http.StatusCreated)
	svcID := idOf(t, rec)

	mapping := func(method string) uint {
		r := admin(http.MethodPost, "/admin/proto-mappings", map[string]interface{}{
			"ServiceID": svcID, "RPCMethod": method, "ServiceName": "AuthService", "ProtoPackage": "auth",
		})
		mustCode(t, r, http.StatusCreated)
		return idOf(t, r)
	}
	checkID, otpID := mapping("CheckPhone"), mapping("SendOTP")

	for path, mid := range map[string]uint{"/grpc/check": checkID, "/grpc/otp": otpID} {
		mustCode(t, admin(http.MethodPost, "/admin/routes", map[string]interface{}{
			"Path": path, "Method": "POST", "ServiceID": svcID, "ProtoMappingID": mid, "EndpointFilter": "g" + path,
		}), http.StatusCreated)
	}

	// Two routes on one service reach two different RPC methods.
	got := do(http.MethodPost, "/grpc/check", map[string]string{"phone_number": "0812"}, map[string]string{"Authorization": "Bearer abc", "X-Request-Id": "rid-1"})
	mustCode(t, got, http.StatusOK)
	if !bytes.Contains(got.Body.Bytes(), []byte("checked")) {
		t.Fatalf("unexpected CheckPhone body: %s", got.Body.String())
	}
	got = do(http.MethodPost, "/grpc/otp", map[string]string{"phone_number": "0812"}, nil)
	mustCode(t, got, http.StatusOK)
	if !bytes.Contains(got.Body.Bytes(), []byte("otp-sent")) {
		t.Fatalf("unexpected SendOTP body: %s", got.Body.String())
	}

	// JSON field and Authorization header reached the upstream.
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.gotPhone != "0812" {
		t.Errorf("phone = %q", fake.gotPhone)
	}
	if v := fake.gotMeta.Get("authorization"); len(v) != 1 || v[0] != "Bearer abc" {
		t.Errorf("authorization metadata = %v", v)
	}
	if v := fake.gotMeta.Get("x-request-id"); len(v) != 1 || v[0] != "rid-1" {
		t.Errorf("x-request-id metadata = %v", v)
	}
}
