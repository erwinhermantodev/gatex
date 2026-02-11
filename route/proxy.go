package route

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

type GenericProxyHandler struct {
	service database.Service
}

func NewGenericProxyHandler(service database.Service) *GenericProxyHandler {
	return &GenericProxyHandler{service: service}
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

func (h *GenericProxyHandler) handleGRPC(c echo.Context) error {
	db := database.GetDB()
	var mapping database.ProtoMapping
	if err := db.Where("service_id = ?", h.service.ID).First(&mapping).Error; err != nil {
		tracing.Error(c.Request().Context(), "gRPC", "No proto mapping found")
		return echo.NewHTTPError(http.StatusNotFound, "gRPC mapping not found for this service")
	}

	tracing.Info(c.Request().Context(), "gRPC", "Dialing "+h.service.GRPCAddr)
	conn, err := grpc.Dial(h.service.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		tracing.Error(c.Request().Context(), "gRPC", "Dial failed: "+err.Error())
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Failed to connect to gRPC service")
	}
	defer conn.Close()

	ctx := c.Request().Context()
	client := grpcreflect.NewClient(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(conn))
	defer client.Reset()

	fullServiceName := fmt.Sprintf("%s.%s", mapping.ProtoPackage, mapping.ServiceName)
	svcDesc, err := client.ResolveService(fullServiceName)
	if err != nil {
		tracing.Error(ctx, "gRPC", "Service resolution failed: "+err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to resolve gRPC service: %v", err))
	}

	methodDesc := svcDesc.FindMethodByName(mapping.RPCMethod)
	if methodDesc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "gRPC method not found")
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	reqMsg := dynamic.NewMessage(methodDesc.GetInputType())
	if err := json.Unmarshal(body, reqMsg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Failed to parse JSON into gRPC request: %v", err))
	}

	resMsg := dynamic.NewMessage(methodDesc.GetOutputType())
	tracing.Info(ctx, "gRPC", "Invoking method "+mapping.RPCMethod)
	err = conn.Invoke(ctx, fmt.Sprintf("/%s/%s", fullServiceName, mapping.RPCMethod), reqMsg, resMsg)
	if err != nil {
		tracing.Error(ctx, "gRPC", "Invocation failed: "+err.Error())
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.Unavailable || s.Code() == codes.Internal {
				return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("gRPC call failed: %v", err))
			}
		}
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
