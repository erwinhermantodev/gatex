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

	// 3. Dispatch to handler or generic proxy
	finalHandler := h.resolveHandler(c, dbRoute)

	// 4. Apply resilience middlewares dynamically from DB if any
	// Note: Standard route middlewares are applied in route.go,
	// but we can add more "organic" ones here or let them be part of the route MW.
	return finalHandler(c)
}

func (h *DynamicHandler) resolveHandler(c echo.Context, dbRoute database.Route) echo.HandlerFunc {
	// Check for specifically implemented handlers first
	if handler, ok := endpoint[dbRoute.EndpointFilter]; ok {
		return handler.Handle
	}

	// Fallback to Generic Proxy
	proxy := NewGenericProxyHandler(dbRoute.Service)
	return proxy.Handle
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
