package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGrpc processes the gRPC request generically
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
	result, err := h.clientFunc(ctx, h.client, req)
	if err != nil {
		// Handle gRPC specific errors
		return h.handleGRPCError(c, err)
	}

	// Build response safely
	resp, err := h.buildResponseGRPC(result)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process response")
	}

	return c.JSON(http.StatusOK, resp)
}

// handleGRPCError handles gRPC specific errors and maps them to appropriate HTTP status codes
func (h *AuthHandler) handleGRPCError(c echo.Context, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error, check for common client errors in the message
		errorMsg := err.Error()
		if contains(errorMsg, "invalid") || contains(errorMsg, "validation") {
			return echo.NewHTTPError(http.StatusBadRequest, h.operation+" failed: "+errorMsg)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, h.operation+" failed")
	}

	// Map gRPC codes to HTTP status codes
	switch st.Code() {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return echo.NewHTTPError(http.StatusBadRequest, st.Message())
	case codes.Unauthenticated:
		return echo.NewHTTPError(http.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		return echo.NewHTTPError(http.StatusForbidden, st.Message())
	case codes.NotFound:
		return echo.NewHTTPError(http.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		return echo.NewHTTPError(http.StatusConflict, st.Message())
	case codes.ResourceExhausted:
		return echo.NewHTTPError(http.StatusTooManyRequests, st.Message())
	case codes.Canceled:
		return echo.NewHTTPError(499, st.Message()) // Client Closed Request
	case codes.DeadlineExceeded:
		return echo.NewHTTPError(http.StatusGatewayTimeout, st.Message())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, h.operation+" failed: "+st.Message())
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
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

// Factory functions for creating specific handlers using gRPC client

func createGrpcHandler(clientFunc func(ctx context.Context, client client.AuthClient, request interface{}) (map[string]interface{}, error), requestFactory func() interface{}, operation string) *AuthHandler {
	return &AuthHandler{
		client:      client.NewGrpcAuthClient(),
		clientFunc:  clientFunc,
		requestType: requestFactory,
		operation:   operation,
	}
}

func NewLoginHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.Login(ctx, req.(*auth.LoginRequest))
	}, func() interface{} { return &auth.LoginRequest{} }, "Authentication")
}

func NewCheckPhoneHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.CheckPhone(ctx, req.(*auth.CheckPhoneRequest))
	}, func() interface{} { return &auth.CheckPhoneRequest{} }, "Phone check")
}

func NewRefreshTokenHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.RefreshToken(ctx, req.(*auth.RefreshTokenRequest))
	}, func() interface{} { return &auth.RefreshTokenRequest{} }, "Token refresh")
}

func NewLogoutHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.Logout(ctx, req.(*auth.RefreshTokenRequest))
	}, func() interface{} { return &auth.RefreshTokenRequest{} }, "Logout")
}

func NewActivationInitiateHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.ActivationInitiate(ctx, req.(*auth.ActivationRequest))
	}, func() interface{} { return &auth.ActivationRequest{} }, "Activation initiate")
}

func NewActivationCompleteHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.ActivationComplete(ctx, req.(*auth.ActivationRequest))
	}, func() interface{} { return &auth.ActivationRequest{} }, "Activation complete")
}

func NewOtpSendHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.OtpSend(ctx, req.(*auth.OtpRequest))
	}, func() interface{} { return &auth.OtpRequest{} }, "OTP send")
}

func NewOtpVerifyHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.OtpVerify(ctx, req.(*auth.OtpRequest))
	}, func() interface{} { return &auth.OtpRequest{} }, "OTP verify")
}

func NewRegisterRequestHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.RegisterRequest(ctx, req.(*auth.LoginRequest))
	}, func() interface{} { return &auth.LoginRequest{} }, "Register request")
}

func NewRegisterCompleteHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.RegisterComplete(ctx, req.(*auth.ActivationRequest))
	}, func() interface{} { return &auth.ActivationRequest{} }, "Register complete")
}

func NewProfileHandlerGRPC() *AuthHandler {
	return createGrpcHandler(func(ctx context.Context, c client.AuthClient, req interface{}) (map[string]interface{}, error) {
		return c.Profile(ctx, req.(*auth.LoginRequest))
	}, func() interface{} { return &auth.LoginRequest{} }, "Profile")
}
