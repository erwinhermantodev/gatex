package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth"
	rest "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/client"
)

var (
	loadEnvOnce sync.Once
	envLoadErr  error
)

func LoadEnv() error {
	loadEnvOnce.Do(func() {
		envLoadErr = godotenv.Load()
	})
	if envLoadErr != nil {
		return fmt.Errorf("error loading .env file: %v", envLoadErr)
	}
	return nil
}

type InitClient struct {
	Client  *rest.RestClient
	BaseURL string
	ApiKey  string
}

func InitializeClient() (InitClient, error) {
	baseURL := os.Getenv("AUTH_SERVICE_BASE_URL")
	apiKey := ""
	if baseURL == "" {
		return InitClient{}, errors.New("environment variables AUTH_SERVICE_BASE_URL not set")
	}

	client := rest.NewRestClient(apiKey)
	return InitClient{
		Client:  client,
		BaseURL: baseURL,
		ApiKey:  apiKey,
	}, nil
}

func performRequest(ctx context.Context, endpoint string, Method string, payload map[string]interface{}) (map[string]interface{}, error) {
	client, err := InitializeClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %v", err)
	}

	url := fmt.Sprintf("%s%s", client.BaseURL, endpoint)

	resClient, err := client.Client.CallAPI(ctx, Method, url, payload)
	if err != nil {
		return nil, err
	}

	return resClient, nil
}

func Login(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
		"password":    request.Password,
	}
	return performRequest(ctx, "/login", "POST", payload)
}

func CheckPhone(ctx context.Context, request *auth.CheckPhoneRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
	}
	return performRequest(ctx, "/check-phone", "POST", payload)
}

func RefreshToken(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"refreshToken": request.RefreshToken,
	}
	return performRequest(ctx, "/refresh-token", "POST", payload)
}

func Logout(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"refreshToken": request.RefreshToken,
	}
	return performRequest(ctx, "/logout", "POST", payload)
}

func ActivationInitiate(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
		"accountNo":   request.AccountNo,
		"nik":         request.NIK,
		"birthDate":   request.BirthDate,
		"motherName":  request.MotherName,
	}
	return performRequest(ctx, "/activation/initiate", "POST", payload)
}

func ActivationComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

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
	return performRequest(ctx, "/activation/complete", "POST", payload)
}

func OtpSend(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
	}
	return performRequest(ctx, "/otp/send", "POST", payload)
}

func OtpVerify(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
		"otpCode":     request.OtpCode,
	}
	return performRequest(ctx, "/otp/verify", "POST", payload)
}

func RegisterRequest(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
	}
	return performRequest(ctx, "/register/request", "POST", payload)
}

func RegisterComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	// Create payload matching the curl request structure
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

	return performRequest(ctx, "/register/complete", "POST", payload)
}

func Profile(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{}
	return performRequest(ctx, "/profile", "GET", payload)
}
