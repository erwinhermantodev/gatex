package domain

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
