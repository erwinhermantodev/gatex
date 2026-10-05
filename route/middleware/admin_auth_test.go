package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAdminAuth(t *testing.T) {
	ok := func(c echo.Context) error { return c.NoContent(http.StatusOK) }
	cases := []struct {
		name, token, header string
		want                int
	}{
		{"unset token fails closed", "", "Bearer x", http.StatusServiceUnavailable},
		{"missing header", "s3cret", "", http.StatusUnauthorized},
		{"wrong token", "s3cret", "Bearer nope", http.StatusUnauthorized},
		{"valid token", "s3cret", "Bearer s3cret", http.StatusOK},
	}
	for _, c := range cases {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/admin/services", nil)
		if c.header != "" {
			req.Header.Set(echo.HeaderAuthorization, c.header)
		}
		rec := httptest.NewRecorder()
		err := AdminAuth(c.token)(ok)(e.NewContext(req, rec))
		got := rec.Code
		if he, isHTTP := err.(*echo.HTTPError); isHTTP {
			got = he.Code
		}
		if got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
}
