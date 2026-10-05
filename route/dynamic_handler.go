package route

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
)

// DynamicHandler dispatches a request for one route (with its service
// preloaded at registry load time) to a hand-written handler or the generic proxy.
type DynamicHandler struct {
	route database.Route
}

func NewDynamicHandler(route database.Route) *DynamicHandler {
	return &DynamicHandler{route: route}
}

func (h *DynamicHandler) Handle(c echo.Context) error {
	// Specifically implemented handlers take precedence over the generic proxy.
	if handler, ok := endpoint[h.route.EndpointFilter]; ok {
		return handler.Handle(c)
	}
	return NewGenericProxyHandler(h.route.Service, h.route.ProtoMapping).Handle(c)
}
