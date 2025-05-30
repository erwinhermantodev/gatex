package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth/client"
)

// ClientFunc represents a generic client function signature
type ClientFunc func(ctx context.Context, request interface{}) (map[string]interface{}, error)

// RequestValidator interface for request validation
type RequestValidator interface {
	Validate() error
}

// AuthHandler is a generic handler that can handle multiple auth endpoints
type AuthHandler struct {
	clientFunc  ClientFunc
	requestType func() interface{} // Factory function to create new request instance
	operation   string             // For error messages
}

// Handle processes the request generically
func (h *AuthHandler) Handle(c echo.Context) error {
	ctx := c.Request().Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Create new request instance
	req := h.requestType()

	// Parse and bind request
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate request using Echo's validator
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Call the client function
	result, err := h.clientFunc(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, h.operation+" failed")
	}

	// Build response safely
	resp, err := h.buildResponse(result)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process response")
	}

	return c.JSON(http.StatusOK, resp)
}

// buildResponse safely builds the response
func (h *AuthHandler) buildResponse(result map[string]interface{}) (*domain.ClientResponse, error) {
	if result == nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Empty response from auth service")
	}

	resp := &domain.ClientResponse{}

	if code, ok := result["code"]; ok {
		if codeStr, ok := code.(string); ok {
			resp.Code = codeStr
		}
	}

	if message, ok := result["message"]; ok {
		if msgStr, ok := message.(string); ok {
			resp.Message = msgStr
		}
	}

	if data, ok := result["data"]; ok {
		resp.Data = data
	}

	return resp, nil
}

// Factory functions for creating specific handlers

func NewLoginHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.Login(ctx, request.(*auth.LoginRequest))
		},
		requestType: func() interface{} { return &auth.LoginRequest{} },
		operation:   "Authentication",
	}
}

func NewCheckPhoneHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.CheckPhone(ctx, request.(*auth.LoginRequest))
		},
		requestType: func() interface{} { return &auth.LoginRequest{} },
		operation:   "Phone check",
	}
}

func NewRefreshTokenHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.RefreshToken(ctx, request.(*auth.LoginRequest))
		},
		requestType: func() interface{} { return &auth.LoginRequest{} },
		operation:   "Token refresh",
	}
}

func NewLogoutHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.Logout(ctx, request.(*auth.RefreshTokenRequest))
		},
		requestType: func() interface{} { return &auth.RefreshTokenRequest{} },
		operation:   "Logout",
	}
}

func NewActivationInitiateHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.ActivationInitiate(ctx, request.(*auth.ActivationRequest))
		},
		requestType: func() interface{} { return &auth.ActivationRequest{} },
		operation:   "Activation initiate",
	}
}

func NewActivationCompleteHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.ActivationComplete(ctx, request.(*auth.ActivationRequest))
		},
		requestType: func() interface{} { return &auth.ActivationRequest{} },
		operation:   "Activation complete",
	}
}

func NewOtpSendHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.OtpSend(ctx, request.(*auth.OtpRequest))
		},
		requestType: func() interface{} { return &auth.OtpRequest{} },
		operation:   "OTP send",
	}
}

func NewOtpVerifyHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.OtpVerify(ctx, request.(*auth.OtpRequest))
		},
		requestType: func() interface{} { return &auth.OtpRequest{} },
		operation:   "OTP verify",
	}
}

func NewRegisterRequestHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.RegisterRequest(ctx, request.(*auth.LoginRequest))
		},
		requestType: func() interface{} { return &auth.LoginRequest{} },
		operation:   "Register request",
	}
}

func NewRegisterCompleteHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.RegisterComplete(ctx, request.(*auth.ActivationRequest))
		},
		requestType: func() interface{} { return &auth.ActivationRequest{} },
		operation:   "Register complete",
	}
}

func NewProfileHandler() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.Profile(ctx, request.(*auth.LoginRequest))
		},
		requestType: func() interface{} { return &auth.LoginRequest{} },
		operation:   "Profile",
	}
}
