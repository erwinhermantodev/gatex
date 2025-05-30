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
	// Authentication endpoints
	"login":         auth.NewLoginHandler(),
	"check-phone":   auth.NewCheckPhoneHandler(),
	"refresh-token": auth.NewRefreshTokenHandler(),
	"logout":        auth.NewLogoutHandler(),

	// Activation endpoints
	"activation-initiate": auth.NewActivationInitiateHandler(),
	"activation-complete": auth.NewActivationCompleteHandler(),

	// OTP endpoints
	"otp-send":   auth.NewOtpSendHandler(),
	"otp-verify": auth.NewOtpVerifyHandler(),

	// Registration endpoints
	"register-request":  auth.NewRegisterRequestHandler(),
	"register-complete": auth.NewRegisterCompleteHandler(),

	// Profile endpoint
	"profile": auth.NewProfileHandler(),
}
