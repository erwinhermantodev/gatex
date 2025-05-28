package route

import (
	"github.com/labstack/echo/v4"
	auth "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth/handler"
)

// Handler endpoint to use it later
type Handler interface {
	Handle(c echo.Context) (err error)
}

var endpoint = map[string]Handler{
	"login": auth.NewLogin(),
}
