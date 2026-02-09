package domain

import "github.com/go-playground/validator/v10"

type ClientResponse struct {
	Status  bool        `json:"status"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(code, message string, data interface{}) *ClientResponse {
	success := true
	return &ClientResponse{
		Status:  success,
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(code, message string) *ClientResponse {
	success := false
	return &ClientResponse{
		Status:  success,
		Code:    code,
		Message: message,
	}
}

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if cv.Validator == nil {
		return nil
	}
	return cv.Validator.Struct(i)
}
