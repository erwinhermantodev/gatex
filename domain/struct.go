package domain

type ClientResponse struct {
	Status  bool        `json:"status"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
