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
// type ClientFunc func(ctx context.Context, request interface{}) (map[string]interface{}, error)

// // RequestValidator interface for request validation
// type RequestValidator interface {
// 	Validate() error
// }

// // AuthHandler is a generic handler that can handle multiple auth endpoints
// type AuthHandler struct {
// 	clientFunc  ClientFunc
// 	requestType func() interface{} // Factory function to create new request instance
// 	operation   string             // For error messages
// }

// Handle processes the request generically
func (h *AuthHandler) HandleGrpc(c echo.Context) error {
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
		// Handle gRPC specific errors
		return h.handleGRPCError(c, err)
	}

	// Build response safely
	resp, err := h.buildResponseGRPC(result)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process response")
	}

	// Return appropriate HTTP status based on success
	statusCode := http.StatusOK
	// if resp.Status != nil && !*resp.Status {
	// 	statusCode = http.StatusBadRequest
	// }

	return c.JSON(statusCode, resp)
}

// handleGRPCError handles gRPC specific errors and maps them to appropriate HTTP status codes
func (h *AuthHandler) handleGRPCError(c echo.Context, err error) error {
	// You can add more sophisticated gRPC error handling here
	// For example, mapping specific gRPC status codes to HTTP status codes

	// For now, treat all gRPC errors as internal server errors
	// unless they contain specific messages that indicate client errors
	errorMsg := err.Error()

	// Check for common client errors
	if contains(errorMsg, "invalid") || contains(errorMsg, "validation") {
		return echo.NewHTTPError(http.StatusBadRequest, h.operation+" failed: "+errorMsg)
	}

	if contains(errorMsg, "unauthorized") || contains(errorMsg, "forbidden") {
		return echo.NewHTTPError(http.StatusUnauthorized, h.operation+" failed: "+errorMsg)
	}

	if contains(errorMsg, "not found") {
		return echo.NewHTTPError(http.StatusNotFound, h.operation+" failed: "+errorMsg)
	}

	// Default to internal server error
	return echo.NewHTTPError(http.StatusInternalServerError, h.operation+" failed")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// buildResponseGRPC safely builds the response
func (h *AuthHandler) buildResponseGRPC(result map[string]interface{}) (*domain.ClientResponse, error) {
	if result == nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Empty response from auth service")
	}

	resp := &domain.ClientResponse{}

	// Handle success field
	if success, ok := result["success"]; ok {
		if successBool, ok := success.(bool); ok {
			resp.Status = successBool
		}
	}

	// Handle code field
	if code, ok := result["code"]; ok {
		if codeStr, ok := code.(string); ok {
			resp.Code = codeStr
		}
	}

	// Handle message field
	if message, ok := result["message"]; ok {
		if msgStr, ok := message.(string); ok {
			resp.Message = msgStr
		}
	}

	// Handle data field
	if data, ok := result["data"]; ok {
		resp.Data = data
	}

	return resp, nil
}

// Factory functions for creating specific handlers

func NewLoginHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcLogin(ctx, request.(*auth.LoginRequest))
		},
		requestType: func() interface{} { return &auth.LoginRequest{} },
		operation:   "Authentication",
	}
}

func NewCheckPhoneHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcCheckPhone(ctx, request.(*auth.CheckPhoneRequest))
		},
		requestType: func() interface{} { return &auth.CheckPhoneRequest{} },
		operation:   "Phone check",
	}
}

func NewRefreshTokenHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcRefreshToken(ctx, request.(*auth.RefreshTokenRequest))
		},
		requestType: func() interface{} { return &auth.RefreshTokenRequest{} },
		operation:   "Token refresh",
	}
}

func NewLogoutHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcLogout(ctx, request.(*auth.RefreshTokenRequest))
		},
		requestType: func() interface{} { return &auth.RefreshTokenRequest{} },
		operation:   "Logout",
	}
}

func NewActivationInitiateHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcActivationInitiate(ctx, request.(*auth.ActivationRequest))
		},
		requestType: func() interface{} { return &auth.ActivationRequest{} },
		operation:   "Activation initiate",
	}
}

func NewActivationCompleteHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcActivationComplete(ctx, request.(*auth.ActivationRequest))
		},
		requestType: func() interface{} { return &auth.ActivationRequest{} },
		operation:   "Activation complete",
	}
}

func NewOtpSendHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcOtpSend(ctx, request.(*auth.OtpRequest))
		},
		requestType: func() interface{} { return &auth.OtpRequest{} },
		operation:   "OTP send",
	}
}

func NewOtpVerifyHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcOtpVerify(ctx, request.(*auth.OtpRequest))
		},
		requestType: func() interface{} { return &auth.OtpRequest{} },
		operation:   "OTP verify",
	}
}

func NewRegisterRequestHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcRegisterRequest(ctx, request.(*auth.LoginRequest))
		},
		requestType: func() interface{} { return &auth.LoginRequest{} },
		operation:   "Register request",
	}
}

func NewRegisterCompleteHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcRegisterComplete(ctx, request.(*auth.ActivationRequest))
		},
		requestType: func() interface{} { return &auth.ActivationRequest{} },
		operation:   "Register complete",
	}
}

func NewProfileHandlerGRPC() *AuthHandler {
	return &AuthHandler{
		clientFunc: func(ctx context.Context, request interface{}) (map[string]interface{}, error) {
			return client.GrpcProfile(ctx, request.(*auth.LoginRequest))
		},
		requestType: func() interface{} { return &auth.LoginRequest{} },
		operation:   "Profile",
	}
}
