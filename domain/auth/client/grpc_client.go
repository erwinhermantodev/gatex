package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/config"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth"
	pb "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/proto/auth"
	util "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/client"
	"google.golang.org/grpc"
)

type GrpcAuthClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

var (
	grpcInstance *GrpcAuthClient
	grpcOnce     sync.Once
)

func NewGrpcAuthClient() *GrpcAuthClient {
	grpcOnce.Do(func() {
		cfg := config.Load()
		if cfg.AuthServiceGRPCAddr == "" {
			log.Println("Warning: AUTH_SERVICE_GRPC_ADDR not set")
			return
		}

		conn := util.Dial(cfg.AuthServiceGRPCAddr)
		client := pb.NewAuthServiceClient(conn)
		grpcInstance = &GrpcAuthClient{
			client: client,
			conn:   conn,
		}
	})
	return grpcInstance
}

func (c *GrpcAuthClient) Login(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
	req := &pb.LoginRequest{
		PhoneNumber: request.PhoneNumber,
		Password:    request.Password,
		Lang:        cfg.DefaultLang,
	}

	resp, err := c.client.Login(ctx, req)
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

func (c *GrpcAuthClient) CheckPhone(ctx context.Context, request *auth.CheckPhoneRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
	req := &pb.CheckPhoneRequest{
		PhoneNumber: request.PhoneNumber,
		Lang:        cfg.DefaultLang,
	}

	resp, err := c.client.CheckPhone(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("check phone failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

func (c *GrpcAuthClient) RefreshToken(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
	req := &pb.RefreshTokenRequest{
		RefreshToken: request.RefreshToken,
		Lang:         cfg.DefaultLang,
	}

	resp, err := c.client.RefreshToken(ctx, req)
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

func (c *GrpcAuthClient) Logout(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
	req := &pb.LogoutRequest{
		RefreshToken: request.RefreshToken,
		Lang:         cfg.DefaultLang,
	}

	resp, err := c.client.Logout(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("logout failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

func (c *GrpcAuthClient) ActivationInitiate(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
	req := &pb.InitiateActivationRequest{
		PhoneNumber: request.PhoneNumber,
		AccountNo:   request.AccountNo,
		Nik:         request.NIK,
		BirthDate:   request.BirthDate,
		MotherName:  request.MotherName,
		Lang:        cfg.DefaultLang,
	}

	resp, err := c.client.InitiateActivation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("activation initiate failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

func (c *GrpcAuthClient) ActivationComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
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
		Lang:         cfg.DefaultLang,
	}

	resp, err := c.client.CompleteActivation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("activation complete failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

func (c *GrpcAuthClient) OtpSend(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
	req := &pb.SendOTPRequest{
		PhoneNumber: request.PhoneNumber,
		Lang:        cfg.DefaultLang,
	}

	resp, err := c.client.SendOTP(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("send OTP failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

func (c *GrpcAuthClient) OtpVerify(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
	req := &pb.VerifyOTPRequest{
		PhoneNumber: request.PhoneNumber,
		OtpCode:     request.OtpCode,
		Lang:        cfg.DefaultLang,
	}

	resp, err := c.client.VerifyOTP(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("verify OTP failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

func (c *GrpcAuthClient) RegisterRequest(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
	req := &pb.RegisterRequest{
		PhoneNumber: request.PhoneNumber,
		Lang:        cfg.DefaultLang,
	}

	resp, err := c.client.Register(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("register request failed: %v", err)
	}

	return map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
		"code":    resp.Code,
	}, nil
}

func (c *GrpcAuthClient) RegisterComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	if c == nil {
		return nil, errors.New("gRPC client not initialized")
	}

	cfg := config.Load()
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
		Lang:         cfg.DefaultLang,
	}

	resp, err := c.client.CompleteRegistration(ctx, req)
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

func (c *GrpcAuthClient) Profile(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	return map[string]interface{}{
		"success": false,
		"message": "Profile endpoint not implemented in gRPC service",
		"code":    "NOT_IMPLEMENTED",
	}, errors.New("profile endpoint not implemented in gRPC service")
}
