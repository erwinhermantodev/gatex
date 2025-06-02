package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"google.golang.org/grpc"

	pb "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/proto/auth" // Adjust this import path to match your proto package

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth"
	util "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/client"
)

var (
	// loadEnvOnce   sync.Once
	// envLoadErr    errorRefreshToken
	grpcClient    pb.AuthServiceClient
	grpcConn      *grpc.ClientConn
	clientOnce    sync.Once
	clientInitErr error
)

// func LoadEnv() error {
// 	loadEnvOnce.Do(func() {
// 		envLoadErr = godotenv.Load()
// 	})
// 	if envLoadErr != nil {
// 		return fmt.Errorf("error loading .env file: %v", envLoadErr)
// 	}
// 	return nil
// }

// InitializeGRPCClient initializes the gRPC client connection
func InitializeGRPCClient() (pb.AuthServiceClient, error) {
	clientOnce.Do(func() {
		authServiceAddr := os.Getenv("AUTH_SERVICE_GRPC_ADDR")
		if authServiceAddr == "" {
			clientInitErr = errors.New("environment variable AUTH_SERVICE_GRPC_ADDR not set")
			return
		}

		// Use the utility gRPC client dialer
		grpcConn = util.Dial(authServiceAddr)
		grpcClient = pb.NewAuthServiceClient(grpcConn)
	})

	if clientInitErr != nil {
		return nil, clientInitErr
	}

	return grpcClient, nil
}

// CloseGRPCConnection closes the gRPC connection
func CloseGRPCConnection() error {
	if grpcConn != nil {
		return grpcConn.Close()
	}
	return nil
}

// Helper function to get default language
func getDefaultLang() string {
	lang := os.Getenv("DEFAULT_LANG")
	if lang == "" {
		return "id" // Default to Indonesian
	}
	return lang
}

// Login authenticates a user
func GrpcLogin(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}

	req := &pb.LoginRequest{
		PhoneNumber: request.PhoneNumber,
		Password:    request.Password,
		Lang:        getDefaultLang(),
	}

	resp, err := client.Login(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}

	result := map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}

	if resp.Data != nil {
		result["data"] = map[string]interface{}{
			"accessToken":  resp.Data.AccessToken,
			"refreshToken": resp.Data.RefreshToken,
		}
	}

	return result, nil
}

// CheckPhone validates a phone number
func GrpcCheckPhone(ctx context.Context, request *auth.CheckPhoneRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}

	req := &pb.CheckPhoneRequest{
		PhoneNumber: request.PhoneNumber,
		Lang:        getDefaultLang(),
	}

	resp, err := client.CheckPhone(ctx, req)
	if err != nil {
		log.Println("err")
		log.Println(err)
		return nil, fmt.Errorf("check phone failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

// RefreshToken refreshes an access token
func GrpcRefreshToken(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}

	req := &pb.RefreshTokenRequest{
		RefreshToken: request.RefreshToken,
		Lang:         getDefaultLang(),
	}

	resp, err := client.RefreshToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("refresh token failed: %v", err)
	}

	result := map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}

	if resp.Data != nil {
		result["data"] = map[string]interface{}{
			"accessToken":  resp.Data.AccessToken,
			"refreshToken": resp.Data.RefreshToken,
		}
	}

	return result, nil
}

// Logout logs out a user
func GrpcLogout(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	log.Println("InitializeGRPCClient")
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}
	req := &pb.LogoutRequest{
		RefreshToken: request.RefreshToken,
		Lang:         getDefaultLang(),
	}

	resp, err := client.Logout(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("logout failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

// ActivationInitiate initiates account activation
func GrpcActivationInitiate(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}

	req := &pb.InitiateActivationRequest{
		PhoneNumber: request.PhoneNumber,
		AccountNo:   request.AccountNo,
		Nik:         request.NIK,
		BirthDate:   request.BirthDate,
		MotherName:  request.MotherName,
		Lang:        getDefaultLang(),
	}

	resp, err := client.InitiateActivation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("activation initiate failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

// ActivationComplete completes account activation
func GrpcActivationComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}

	req := &pb.CompleteActivationRequest{
		PhoneNumber:  request.PhoneNumber,
		Password:     request.Password,
		FullName:     request.FullName,
		NickName:     request.NickName,
		BirthPlace:   request.BirthPlace,
		BirthDate:    request.BirthDate,
		Gender:       request.Gender,
		Religion:     request.Religion,
		Address:      request.Address,
		Rt:           request.RT,
		Rw:           request.RW,
		Province:     request.Province,
		City:         request.City,
		District:     request.District,
		SubDistrict:  request.SubDistrict,
		PostalCode:   request.PostalCode,
		Npwp:         request.NPWP,
		Email:        request.Email,
		Occupation:   request.Occupation,
		FundPurpose:  request.FundPurpose,
		FundSource:   request.FundSource,
		AnnualIncome: request.AnnualIncome,
		Lang:         getDefaultLang(),
	}

	resp, err := client.CompleteActivation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("activation complete failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

// OtpSend sends an OTP
func GrpcOtpSend(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}

	req := &pb.SendOTPRequest{
		PhoneNumber: request.PhoneNumber,
		Lang:        getDefaultLang(),
	}

	resp, err := client.SendOTP(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("send OTP failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

// OtpVerify verifies an OTP
func GrpcOtpVerify(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}

	req := &pb.VerifyOTPRequest{
		PhoneNumber: request.PhoneNumber,
		OtpCode:     request.OtpCode,
		Lang:        getDefaultLang(),
	}

	resp, err := client.VerifyOTP(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("verify OTP failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

// RegisterRequest initiates user registration
func GrpcRegisterRequest(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}

	req := &pb.RegisterRequest{
		PhoneNumber: request.PhoneNumber,
		Lang:        getDefaultLang(),
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("register request failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

// RegisterComplete completes user registration
func GrpcRegisterComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	client, err := InitializeGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC client: %v", err)
	}

	req := &pb.CompleteRegistrationRequest{
		PhoneNumber:  request.PhoneNumber,
		Password:     request.Password,
		ReferralCode: request.ReferralCode,
		FullName:     request.FullName,
		NickName:     request.NickName,
		BirthPlace:   request.BirthPlace,
		BirthDate:    request.BirthDate,
		Gender:       request.Gender,
		Religion:     request.Religion,
		Address:      request.Address,
		Rt:           request.RT,
		Rw:           request.RW,
		Province:     request.Province,
		City:         request.City,
		District:     request.District,
		SubDistrict:  request.SubDistrict,
		PostalCode:   request.PostalCode,
		Npwp:         request.NPWP,
		Email:        request.Email,
		Occupation:   request.Occupation,
		FundPurpose:  request.FundPurpose,
		FundSource:   request.FundSource,
		AnnualIncome: request.AnnualIncome,
		Lang:         getDefaultLang(),
	}

	resp, err := client.CompleteRegistration(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("register complete failed: %v", err)
	}

	result := map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}

	if resp.Data != nil {
		result["data"] = map[string]interface{}{
			"accessToken":  resp.Data.AccessToken,
			"refreshToken": resp.Data.RefreshToken,
		}
	}

	return result, nil
}

// Profile gets user profile (this would need to be implemented in your auth service)
func GrpcProfile(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	// Note: Profile endpoint is not defined in the protobuf service
	// You might need to add this to your auth service protobuf definition
	// For now, returning a placeholder response
	return map[string]interface{}{
		"success": false,
		"message": "Profile endpoint not implemented in gRPC service",
		"code":    "NOT_IMPLEMENTED",
	}, errors.New("profile endpoint not implemented in gRPC service")
}
