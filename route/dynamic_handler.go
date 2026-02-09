package route

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth/client"
)

// DynamicHandler resolves the service and method from DB and dispatches the call
type DynamicHandler struct {
	endpoint string
}

func NewDynamicHandler(endpoint string) *DynamicHandler {
	return &DynamicHandler{endpoint: endpoint}
}

func (h *DynamicHandler) Handle(c echo.Context) error {
	// 1. Get route/endpoint configuration from DB
	db := database.GetDB()
	var dbRoute database.Route
	if err := db.Preload("Service").Where("endpoint_filter = ?", h.endpoint).First(&dbRoute).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Endpoint configuration not found")
	}

	// 2. Resolve the client based on service protocol
	// For now, we still use the AuthClient interface for auth-service
	// In a fully generic gateway, we would use gRPC reflection or dynamic messaging
	if dbRoute.Service.Name == "auth-service" {
		var authClient client.AuthClient
		if dbRoute.Service.Protocol == "grpc" {
			authClient = client.NewGrpcAuthClient()
		} else {
			authClient = client.NewRestAuthClient()
		}

		// Dispatch based on endpoint_filter
		// This is still a bit static, but moves the selection to DB
		// To be fully generic, we'd need a registry of method callers
		return h.dispatchAuthCall(c, authClient, dbRoute.EndpointFilter)
	}

	return echo.NewHTTPError(http.StatusNotImplemented, "Service proxy not implemented")
}

func (h *DynamicHandler) dispatchAuthCall(c echo.Context, authClient client.AuthClient, filter string) error {
	// This part can be further generalized using a map of functions
	// or reflection, but let's keep it simple for now.
	// The routing is already dynamic from DB.

	// If the handler exists in our static map, use it.
	// This maintains backward compatibility with specifically implemented handlers.
	if handler, ok := endpoint[filter]; ok {
		return handler.Handle(c)
	}

	return echo.NewHTTPError(http.StatusNotFound, "Handler implementation not found")
}
