package middleware

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
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

			// We use a background goroutine or just save it synchronously for now
			// Given this is an admin dashboard, sync is fine for low traffic,
			// but goroutine is better for performance.
			go func(l database.RequestLog) {
				db := database.GetDB()
				db.Create(&l)
			}(log)

			return err
		}
	}
}
