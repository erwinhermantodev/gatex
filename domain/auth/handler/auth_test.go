package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth/client"
)

// MockAuthClient is a mock implementation of the AuthClient interface
type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Login(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) CheckPhone(ctx context.Context, request *auth.CheckPhoneRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) Logout(ctx context.Context, request *auth.RefreshTokenRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) ActivationInitiate(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) ActivationComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) OtpSend(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) OtpVerify(ctx context.Context, request *auth.OtpRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) RegisterRequest(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) RegisterComplete(ctx context.Context, request *auth.ActivationRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthClient) Profile(ctx context.Context, request *auth.LoginRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestLoginHandler(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = &domain.CustomValidator{Validator: nil} // Simplified for test
	loginJSON := `{"phoneNumber":"08123456789", "password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockClient := new(MockAuthClient)
	expectedRes := map[string]interface{}{
		"code":    "SUCCESS",
		"message": "Login successful",
		"data": map[string]interface{}{
			"accessToken": "test-token",
		},
	}
	mockClient.On("Login", mock.Anything, mock.Anything).Return(expectedRes, nil)

	h := &AuthHandler{
		client: mockClient,
		clientFunc: func(ctx context.Context, client client.AuthClient, request interface{}) (map[string]interface{}, error) {
			return client.Login(ctx, request.(*auth.LoginRequest))
		},
		requestType: func() interface{} { return &auth.LoginRequest{} },
		operation:   "Login",
	}

	// Assertions
	if assert.NoError(t, h.Handle(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "SUCCESS", response["code"])
		assert.Equal(t, "Login successful", response["message"])
	}
}
