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
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type GenericProxyHandler struct {
	service database.Service
}

func NewGenericProxyHandler(service database.Service) *GenericProxyHandler {
	return &GenericProxyHandler{service: service}
}

func (h *GenericProxyHandler) Handle(c echo.Context) error {
	tracing.Info(c.Request().Context(), "Proxy", "Interpreting request for "+h.service.Name)
	if h.service.Protocol == "grpc" {
		return h.handleGRPC(c)
	}

	target, err := url.Parse(h.service.BaseURL)
	if err != nil {
		tracing.Error(c.Request().Context(), "REST", "Invalid upstream URL: "+h.service.BaseURL)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid Upstream URL")
	}

	tracing.Info(c.Request().Context(), "REST", "Proxying to "+h.service.BaseURL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the director to preserve path and handle headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Map the path: strip the gateway prefix if needed
		// For example, if gateway path is /api/v1/user and service is at /user
		// In our DB system, we usually map the exact path.

		// Ensure Host header is set correctly for upstream
		req.Host = target.Host

		// Add X-Forwarded headers
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
	// Find mapping by service ID and route path (or we might need more specific logic)
	// For now, let's assume we can find it by service and path
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
	tracing.Info(ctx, "gRPC", "Resolving service "+fullServiceName)
	svcDesc, err := client.ResolveService(fullServiceName)
	if err != nil {
		tracing.Error(ctx, "gRPC", "Service resolution failed: "+err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to resolve gRPC service: %v", err))
	}

	methodDesc := svcDesc.FindMethodByName(mapping.RPCMethod)
	if methodDesc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "gRPC method not found")
	}

	// Read JSON body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Create dynamic message for request
	reqMsg := dynamic.NewMessage(methodDesc.GetInputType())
	if err := json.Unmarshal(body, reqMsg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Failed to parse JSON into gRPC request: %v", err))
	}

	// Perform call
	resMsg := dynamic.NewMessage(methodDesc.GetOutputType())
	tracing.Info(ctx, "gRPC", "Invoking method "+mapping.RPCMethod)
	err = conn.Invoke(ctx, fmt.Sprintf("/%s/%s", fullServiceName, mapping.RPCMethod), reqMsg, resMsg)
	if err != nil {
		tracing.Error(ctx, "gRPC", "Invocation failed: "+err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("gRPC call failed: %v", err))
	}

	// Convert response back to JSON
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
