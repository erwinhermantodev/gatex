package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/metrics"
)

func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)

		duration := time.Since(start)
		status := c.Response().Status
		path := c.Path()

		// Get service name from context if set by SetContextValue
		service := "unknown"
		if val := c.Get(util.ContextRouterKey); val != nil {
			if s, ok := val.(string); ok {
				service = s
			}
		}

		metrics.Record(service, path, status, duration)

		return err
	}
}
