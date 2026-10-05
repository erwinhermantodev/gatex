package route

import (
	"encoding/json"
	"time"

	"github.com/labstack/echo/v4"
	customMw "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route/middleware"
)

var middlewareHandler = map[string]echo.MiddlewareFunc{
	"timeout":         customMw.TimeoutMiddleware(10 * time.Second),
	"retry":           customMw.RetryMiddleware(3),
	"circuit-breaker": customMw.CircuitBreakerMiddleware("default"),
}

// middlewareNames decodes the JSON array stored in Route.Middleware,
// dropping blank entries.
func middlewareNames(raw string) []string {
	var all, names []string
	_ = json.Unmarshal([]byte(raw), &all)
	for _, n := range all {
		if n != "" {
			names = append(names, n)
		}
	}
	return names
}
