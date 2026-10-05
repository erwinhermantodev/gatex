package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/tracing"
)

func TrafficLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = c.Request().Header.Get(echo.HeaderXRequestID)
			}

			// Attach RequestID to context for tracing
			ctx := context.WithValue(c.Request().Context(), tracing.RequestIDKey, requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			start := time.Now()
			err := next(c)

			// Skip the dashboard and its admin polling to keep the log about real traffic.
			if path := c.Request().URL.Path; util.IsAdminPath(path) || strings.HasPrefix(path, "/dashboard") {
				return err
			}
			latency := time.Since(start)

			// Extract request info
			req := c.Request()
			res := c.Response()

			log := database.RequestLog{
				RequestID:  requestID,
				Method:     req.Method,
				Path:       req.URL.Path,
				StatusCode: res.Status,
				LatencyMS:  latency.Milliseconds(),
				ClientIP:   c.RealIP(),
				UserAgent:  req.UserAgent(),
			}

			if err != nil {
				log.ErrorMessage = err.Error()
			}

			database.EnqueueRequestLog(log)

			return err
		}
	}
}
