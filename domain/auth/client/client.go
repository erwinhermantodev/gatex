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

func performRequest(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	client, err := InitializeClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %v", err)
	}

	url := fmt.Sprintf("%s%s", client.BaseURL, endpoint)

	resClient, err := client.Client.CallAPI(ctx, "POST", url, payload)
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
	return performRequest(ctx, "/login", payload)
}

func CheckPhone(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
	}
	return performRequest(ctx, "check-phone", payload)
}

func RefreshToken(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	payload := map[string]interface{}{
		"phoneNumber": request.PhoneNumber,
	}
	return performRequest(ctx, "refresh-token", payload)
}
