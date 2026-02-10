package route

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
)

type GenericProxyHandler struct {
	service database.Service
}

func NewGenericProxyHandler(service database.Service) *GenericProxyHandler {
	return &GenericProxyHandler{service: service}
}

func (h *GenericProxyHandler) Handle(c echo.Context) error {
	if h.service.Protocol == "grpc" {
		// For a truly generic gRPC proxy, one would use something like
		// grpc-gateway or dynamic gRPC reflection.
		// For now, we return 501 for non-implemented gRPC services.
		return echo.NewHTTPError(http.StatusNotImplemented, "Generic gRPC proxying not yet implemented")
	}

	target, err := url.Parse(h.service.BaseURL)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid Upstream URL")
	}

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

// RegisterHandler registers a handler in the global endpoint map
func RegisterHandler(name string, h Handler) {
	endpoint[name] = h
}
