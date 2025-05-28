package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/errors"
	"google.golang.org/grpc/codes"
)

type RestClient struct {
	APIKey string
}

func NewRestClient(apiKey string) *RestClient {
	return &RestClient{APIKey: apiKey}
}

func (a *RestClient) CallAPI(ctx context.Context, method, url string, payload map[string]interface{}) (map[string]interface{}, error) {
	var req *http.Request
	var err error
	log.Println("CallAPI")
	log.Println("==========================")
	log.Println(method)
	log.Println(url)

	switch method {
	case "GET":
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			log.Println("Error creating GET request:", err)
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	case "POST":
		reqBody, err := json.Marshal(payload)
		if err != nil {
			log.Println("Error marshaling POST payload:", err)
			return nil, err
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
		if err != nil {
			log.Println("Error creating POST request:", err)
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	case "PUT":
		reqBody, err := json.Marshal(payload)
		if err != nil {
			log.Println("Error marshaling PUT payload:", err)
			return nil, err
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
		if err != nil {
			log.Println("Error creating PUT request:", err)
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	case "DELETE":
		reqBody, err := json.Marshal(payload)
		if err != nil {
			log.Println("Error marshaling DELETE payload:", err)
			return nil, err
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
		if err != nil {
			log.Println("Error creating DELETE request:", err)
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	req.Header.Set("x-api-key", a.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Println(resp)
	log.Println("response")
	log.Println("==========================")

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var httpStatus codes.Code = codes.Code(int32(resp.StatusCode))
		return nil, errors.ErrorMap(httpStatus, util.StatusMessage[httpStatus])
	}

	// Read the response body into a byte slice
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}

	if len(bodyBytes) == 0 {
		log.Println("Response body is empty")
		return nil, fmt.Errorf("response body is empty")
	}

	// Decode the byte slice into a map
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.Println("Error decoding response:", err)
		return nil, err
	}

	return result, nil
}
