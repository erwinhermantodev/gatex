package client

import (
	"context"
	"fmt"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/config"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth"
	rest "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/client"
)

type RestAuthClient struct {
	client  *rest.RestClient
	baseURL string
}

func NewRestAuthClient() *RestAuthClient {
	cfg := config.Load()
	return &RestAuthClient{
		client:  rest.NewRestClient(""),
		baseURL: cfg.AuthServiceBaseURL,
	}
}

func (c *RestAuthClient) performRequest(ctx context.Context, endpoint string, Method string, payload map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	return c.client.CallAPI(ctx, Method, url, payload)
}

func (c *RestAuthClient) Login(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
		"password":    request.Password,
	}
	return c.performRequest(ctx, "/login", "POST", payload)
}

func (c *RestAuthClient) CheckPhone(ctx context.Context, request *auth.CheckPhoneRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
	}
	return c.performRequest(ctx, "/check-phone", "POST", payload)
}

func (c *RestAuthClient) RefreshToken(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"refreshToken": request.RefreshToken,
	}
	return c.performRequest(ctx, "/refresh-token", "POST", payload)
}

func (c *RestAuthClient) Logout(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"refreshToken": request.RefreshToken,
	}
	return c.performRequest(ctx, "/logout", "POST", payload)
}

func (c *RestAuthClient) ActivationInitiate(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
		"accountNo":   request.AccountNo,
		"nik":         request.NIK,
		"birthDate":   request.BirthDate,
		"motherName":  request.MotherName,
	}
	return c.performRequest(ctx, "/activation/initiate", "POST", payload)
}

func (c *RestAuthClient) ActivationComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"phoneNumber":  request.PhoneNumber,
		"password":     request.Password,
		"fullName":     request.FullName,
		"nickName":     request.NickName,
		"birthPlace":   request.BirthPlace,
		"birthDate":    request.BirthDate,
		"gender":       request.Gender,
		"religion":     request.Religion,
		"address":      request.Address,
		"rt":           request.RT,
		"rw":           request.RW,
		"province":     request.Province,
		"city":         request.City,
		"district":     request.District,
		"subDistrict":  request.SubDistrict,
		"postalCode":   request.PostalCode,
		"npwp":         request.NPWP,
		"email":        request.Email,
		"occupation":   request.Occupation,
		"fundPurpose":  request.FundPurpose,
		"fundSource":   request.FundSource,
		"annualIncome": request.AnnualIncome,
	}
	return c.performRequest(ctx, "/activation/complete", "POST", payload)
}

func (c *RestAuthClient) OtpSend(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
	}
	return c.performRequest(ctx, "/otp/send", "POST", payload)
}

func (c *RestAuthClient) OtpVerify(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
		"otpCode":     request.OtpCode,
	}
	return c.performRequest(ctx, "/otp/verify", "POST", payload)
}

func (c *RestAuthClient) RegisterRequest(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
	}
	return c.performRequest(ctx, "/register/request", "POST", payload)
}

func (c *RestAuthClient) RegisterComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"phoneNumber":  request.PhoneNumber,
		"password":     request.Password,
		"referralCode": request.ReferralCode,
		"fullName":     request.FullName,
		"nickName":     request.NickName,
		"birthPlace":   request.BirthPlace,
		"birthDate":    request.BirthDate,
		"gender":       request.Gender,
		"religion":     request.Religion,
		"address":      request.Address,
		"rt":           request.RT,
		"rw":           request.RW,
		"province":     request.Province,
		"city":         request.City,
		"district":     request.District,
		"subDistrict":  request.SubDistrict,
		"postalCode":   request.PostalCode,
		"npwp":         request.NPWP,
		"email":        request.Email,
		"occupation":   request.Occupation,
		"fundPurpose":  request.FundPurpose,
		"fundSource":   request.FundSource,
		"annualIncome": request.AnnualIncome,
	}
	return c.performRequest(ctx, "/register/complete", "POST", payload)
}

func (c *RestAuthClient) Profile(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	return c.performRequest(ctx, "/profile", "GET", nil)
}
