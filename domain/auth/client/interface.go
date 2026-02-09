package client

import (
	"context"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth"
)

// AuthClient defines the interface for authentication service interactions
type AuthClient interface {
	Login(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error)
	CheckPhone(ctx context.Context, request *auth.CheckPhoneRequest) (map[string]interface{}, error)
	RefreshToken(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error)
	Logout(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error)
	ActivationInitiate(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error)
	ActivationComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error)
	OtpSend(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error)
	OtpVerify(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error)
	RegisterRequest(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error)
	RegisterComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error)
	Profile(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error)
}
