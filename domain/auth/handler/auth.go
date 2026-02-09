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
	client      client.AuthClient
	clientFunc  func(ctx context.Context, client client.AuthClient, request interface{}) (map[string]interface{}, error)
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
	result, err := h.clientFunc(ctx, h.client, req)
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

// Factory functions for creating specific handlers using REST client

func createRestHandler(clientFunc func(ctx context.Context, client client.AuthClient, request interface{}) (map[string]interface{}, error), requestFactory func() interface{}, operation string) *AuthHandler {
	return &AuthHandler{
		client:      client.NewRestAuthClient(),
		clientFunc:  clientFunc,
		requestType: requestFactory,
		operation:   operation,
	}
}

func NewLoginHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.Login(ctx, req.(*auth.LoginRequest))
	}, func() interface{} { return &auth.LoginRequest{} }, "Authentication")
}

func NewCheckPhoneHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.CheckPhone(ctx, req.(*auth.CheckPhoneRequest))
	}, func() interface{} { return &auth.CheckPhoneRequest{} }, "Phone check")
}

func NewRefreshTokenHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.RefreshToken(ctx, req.(*auth.RefreshTokenRequest))
	}, func() interface{} { return &auth.RefreshTokenRequest{} }, "Token refresh")
}

func NewLogoutHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.Logout(ctx, req.(*auth.RefreshTokenRequest))
	}, func() interface{} { return &auth.RefreshTokenRequest{} }, "Logout")
}

func NewActivationInitiateHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.ActivationInitiate(ctx, req.(*auth.ActivationRequest))
	}, func() interface{} { return &auth.ActivationRequest{} }, "Activation initiate")
}

func NewActivationCompleteHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.ActivationComplete(ctx, req.(*auth.ActivationRequest))
	}, func() interface{} { return &auth.ActivationRequest{} }, "Activation complete")
}

func NewOtpSendHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.OtpSend(ctx, req.(*auth.OtpRequest))
	}, func() interface{} { return &auth.OtpRequest{} }, "OTP send")
}

func NewOtpVerifyHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.OtpVerify(ctx, req.(*auth.OtpRequest))
	}, func() interface{} { return &auth.OtpRequest{} }, "OTP verify")
}

func NewRegisterRequestHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.RegisterRequest(ctx, req.(*auth.LoginRequest))
	}, func() interface{} { return &auth.LoginRequest{} }, "Register request")
}

func NewRegisterCompleteHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.RegisterComplete(ctx, req.(*auth.ActivationRequest))
	}, func() interface{} { return &auth.ActivationRequest{} }, "Register complete")
}

func NewProfileHandler() *AuthHandler {
	return createRestHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.Profile(ctx, req.(*auth.LoginRequest))
	}, func() interface{} { return &auth.LoginRequest{} }, "Profile")
}
