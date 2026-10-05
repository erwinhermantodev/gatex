package route

import (
	"encoding/json"
	"time"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/config"
	customMw "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route/middleware"
)

var middlewareHandler = map[string]echo.MiddlewareFunc{}

// initMiddleware builds the named route middlewares; call before registry.Reload.
func initMiddleware(cfg *config.Config) {
	middlewareHandler = map[string]echo.MiddlewareFunc{
		customMw.NameTimeout:        customMw.TimeoutMiddleware(10 * time.Second),
		customMw.NameRetry:          customMw.RetryMiddleware(3),
		customMw.NameCircuitBreaker: customMw.CircuitBreakerMiddleware("default"),
		customMw.NameJWT: customMw.JWTAuth(customMw.JWTConfig{
			Secret:       cfg.JWTSecret,
			PublicKeyPEM: cfg.JWTPublicKeyPEM,
			Issuer:       cfg.JWTIssuer,
			Audience:     cfg.JWTAudience,
		}),
		customMw.NameAPIKey: customMw.APIKeyAuth(cfg.APIKeys),
	}
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
