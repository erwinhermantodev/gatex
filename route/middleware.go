package route

import (
	"time"

	"github.com/labstack/echo/v4"
	customMw "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route/middleware"
)

var middlewareHandler = map[string]echo.MiddlewareFunc{
	"timeout":         customMw.TimeoutMiddleware(10 * time.Second),
	"retry":           customMw.RetryMiddleware(3),
	"circuit-breaker": customMw.CircuitBreakerMiddleware("default"),
}
