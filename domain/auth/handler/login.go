package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/auth/client"
)

type Login struct{}

func (h *Login) Handle(c echo.Context) error {
	ctx := c.Request().Context()
	if ctx == nil {
		ctx = context.Background()
	}

	r := new(auth.LoginRequest)
	err := h.validate(r, c)
	if err != nil {
		log.Println("validate error : ", err.Error())
		return err
	}

	result, err := client.Login(ctx, r)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result["message"].(string))
	}

	resp, err := h.buildResponse(result)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Login) buildResponse(response map[string]interface{}) (*domain.ClientResponse, error) {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshaling map to JSON:", err)
		return nil, err
	}

	var resp domain.ClientResponse
	err = json.Unmarshal(jsonBytes, &resp)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return nil, err
	}

	return &resp, nil
}

func (h *Login) validate(r *auth.LoginRequest, c echo.Context) error {
	if err := c.Bind(r); err != nil {
		return err
	}

	return c.Validate(r)
}

func NewLogin() *Login {
	return &Login{}
}
