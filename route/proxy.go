package route

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/netguard"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/tracing"
	"google.golang.org/grpc/metadata"
)

// upstreamTransport is shared across requests (connection reuse) and refuses
// to dial blocked addresses such as cloud metadata endpoints.
var upstreamTransport = &http.Transport{
	Proxy:                 nil,
	DialContext:           netguard.DialContext,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost:   20,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 60 * time.Second,
}

type GenericProxyHandler struct {
	service database.Service
	mapping *database.ProtoMapping // nil: fall back to the service's first mapping
}

func NewGenericProxyHandler(service database.Service, mapping *database.ProtoMapping) *GenericProxyHandler {
	return &GenericProxyHandler{service: service, mapping: mapping}
}

func (h *GenericProxyHandler) Handle(c echo.Context) error {
	stats := util.GetHealthStats(h.service.ID)
	if !stats.ShouldAllow() {
		tracing.Error(c.Request().Context(), "Proxy", "Circuit breaker OPEN for "+h.service.Name)
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Service temporarily unavailable (Circuit Breaker OPEN)")
	}

	tracing.Info(c.Request().Context(), "Proxy", "Interpreting request for "+h.service.Name)
	if h.service.Protocol == "grpc" {
		err := h.handleGRPC(c)
		if err != nil {
			stats.RecordFailure()
		} else {
			stats.RecordSuccess()
		}
		return err
	}

	target, err := url.Parse(h.service.BaseURL)
	if err != nil {
		tracing.Error(c.Request().Context(), "REST", "Invalid upstream URL: "+h.service.BaseURL)
		stats.RecordFailure()
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid Upstream URL")
	}

	tracing.Info(c.Request().Context(), "REST", "Proxying to "+h.service.BaseURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = upstreamTransport

	// Capture response to record success/failure
	proxy.ModifyResponse = func(res *http.Response) error {
		if res.StatusCode >= 500 {
			stats.RecordFailure()
		} else {
			stats.RecordSuccess()
		}
		return nil
	}

	proxy.ErrorHandler = func(res http.ResponseWriter, req *http.Request, err error) {
		stats.RecordFailure()
		tracing.Error(c.Request().Context(), "REST", "Proxy error: "+err.Error())
		c.Error(echo.NewHTTPError(http.StatusBadGateway, "Proxy error"))
	}

	// Customize the director to preserve path and handle headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		if clientIP := c.RealIP(); clientIP != "" {
			req.Header.Set("X-Forwarded-For", clientIP)
		}
	}

	proxy.ServeHTTP(c.Response(), c.Request())
	return nil
}

// grpcSkipHeaders are not forwarded as gRPC metadata.
var grpcSkipHeaders = map[string]bool{
	"host": true, "content-length": true, "content-type": true, "accept-encoding": true,
	"connection": true, "keep-alive": true, "te": true, "trailer": true,
	"transfer-encoding": true, "upgrade": true, "user-agent": true,
}

func grpcMetadata(c echo.Context) metadata.MD {
	md := metadata.MD{}
	for k, vals := range c.Request().Header {
		key := strings.ToLower(k)
		if grpcSkipHeaders[key] || strings.HasPrefix(key, "grpc-") || strings.HasPrefix(key, ":") {
			continue
		}
		md.Append(key, vals...)
	}
	if id := c.Response().Header().Get(echo.HeaderXRequestID); id != "" {
		md.Set("x-request-id", id)
	}
	if ip := c.RealIP(); ip != "" {
		md.Set("x-forwarded-for", ip)
	}
	return md
}

func (h *GenericProxyHandler) handleGRPC(c echo.Context) error {
	ctx := c.Request().Context()

	mapping := h.mapping
	if mapping == nil {
		var m database.ProtoMapping
		if err := database.GetDB().Where("service_id = ?", h.service.ID).First(&m).Error; err != nil {
			tracing.Error(ctx, "gRPC", "No proto mapping found")
			return echo.NewHTTPError(http.StatusNotFound, "gRPC mapping not found for this service")
		}
		mapping = &m
	}

	conn, err := conns.get(h.service.GRPCAddr)
	if err != nil {
		tracing.Error(ctx, "gRPC", "Client setup failed: "+err.Error())
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Failed to connect to gRPC service")
	}

	fullServiceName := fmt.Sprintf("%s.%s", mapping.ProtoPackage, mapping.ServiceName)
	methodDesc, err := resolveMethod(ctx, conn, h.service.GRPCAddr, fullServiceName, mapping.RPCMethod)
	if err != nil {
		tracing.Error(ctx, "gRPC", "Service resolution failed: "+err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("Failed to resolve gRPC service: %v", err))
	}
	if methodDesc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "gRPC method not found")
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	reqMsg := dynamic.NewMessage(methodDesc.GetInputType())
	if len(body) > 0 {
		if err := json.Unmarshal(body, reqMsg); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Failed to parse JSON into gRPC request: %v", err))
		}
	}

	resMsg := dynamic.NewMessage(methodDesc.GetOutputType())
	tracing.Info(ctx, "gRPC", "Invoking method "+mapping.RPCMethod)
	callCtx := metadata.NewOutgoingContext(ctx, grpcMetadata(c))
	err = conn.Invoke(callCtx, fmt.Sprintf("/%s/%s", fullServiceName, mapping.RPCMethod), reqMsg, resMsg)
	if err != nil {
		tracing.Error(ctx, "gRPC", "Invocation failed: "+err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("gRPC call failed: %v", err))
	}

	resJSON, err := resMsg.MarshalJSON()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal gRPC response to JSON")
	}

	return c.JSONBlob(http.StatusOK, resJSON)
}

// RegisterHandler registers a handler in the global endpoint map
func RegisterHandler(name string, h Handler) {
	endpoint[name] = h
}
