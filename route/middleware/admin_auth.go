package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// AdminAuth protects the admin API with a static bearer token.
// It fails closed: when no token is configured every request is rejected.
func AdminAuth(token string) echo.MiddlewareFunc {
	expected := []byte(token)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if len(expected) == 0 {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "Admin API disabled: ADMIN_API_TOKEN is not set")
			}
			got := strings.TrimPrefix(c.Request().Header.Get(echo.HeaderAuthorization), "Bearer ")
			if got == "" {
				got = c.Request().Header.Get("X-Admin-Token")
			}
			if subtle.ConstantTimeCompare([]byte(got), expected) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}
			return next(c)
		}
	}
}
