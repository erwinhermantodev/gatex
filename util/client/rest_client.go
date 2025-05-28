package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"gitlab.com/posfin-unigo/middleware/unigo-nde/api-gateway-go/util"
	"gitlab.com/posfin-unigo/middleware/unigo-nde/api-gateway-go/util/errors"
	"google.golang.org/grpc/codes"
)

type RestClient struct {
	APIKey string
}

func NewRestClient(apiKey string) *RestClient {
	return &RestClient{APIKey: apiKey}
}

func (a *RestClient) CallAPI(method, url string, payload map[string]interface{}) (map[string]interface{}, error) {
	var req *http.Request
	var err error
	log.Println("CallAPI")
	log.Println("==========================")
	log.Println(method)
	log.Println(url)
	switch method {
	case "GET":
		req, err = http.NewRequest(method, url, nil)
		req.Header.Set("Content-Type", "application/json")
	case "POST":
		reqBody, _ := json.Marshal(payload)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
	case "PUT":
		reqBody, _ := json.Marshal(payload)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
	case "DELETE":
		reqBody, _ := json.Marshal(payload)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
	default:
		return nil, fmt.Errorf("unsupported method")
	}

	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
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
