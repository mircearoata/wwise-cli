package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type WwiseClient struct {
	auth string
}

func NewWwiseClient() *WwiseClient {
	return &WwiseClient{}
}

func (client *WwiseClient) Authenticate(email string, password string) error {
	body := map[string]string{"email": email, "password": password}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal auth json: %w", err)
	}

	response, err := http.Post("https://www.audiokinetic.com/wwise/launcher/?action=login", "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		return fmt.Errorf("failed to post auth request: %w", err)
	}

	if response.StatusCode != 200 {
		var responseJson struct {
			Code    int    `json:"code"`
			Err     string `json:"err"`
			ErrCode string `json:"errCode"`
			Random  string `json:"random"`
		}
		err = json.NewDecoder(response.Body).Decode(&responseJson)
		if err != nil {
			return fmt.Errorf("failed to decode auth error response: %w", err)
		}
		return fmt.Errorf("failed to authenticate: %s (error_code: %s)", responseJson.Err, responseJson.ErrCode)
	}

	var responseContent apiResponse
	err = json.NewDecoder(response.Body).Decode(&responseContent)
	if err != nil {
		return fmt.Errorf("failed to decode raw auth response: %w", err)
	}

	responsePayload, err := responseContent.decodePayload()
	if err != nil {
		return fmt.Errorf("failed to decode auth response payload: %w", err)
	}

	var responseJson struct {
		Code          int    `json:"code"`
		Jwt           string `json:"jwt"`
		SpecialAction bool   `json:"specialAction"`
		Random        string `json:"random"`
	}
	err = json.Unmarshal([]byte(responsePayload), &responseJson)
	if err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	client.auth = responseJson.Jwt

	return nil
}

type apiResponse struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

func (res apiResponse) decodePayload() (string, error) {
	payloadBytes, err := base64.StdEncoding.DecodeString(res.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode payload: %w", err)
	}
	return string(payloadBytes), nil
}

func (client *WwiseClient) SendRequest(method string, url string, body interface{}) (string, error) {
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	if !strings.HasPrefix(url, "https://blob-api.gowwise.com") {
		url = "https://blob-api.gowwise.com" + url
	}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+client.auth)
	request.Header.Set("X-client-version", "2025.1.0.5135")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get manifest: %w", err)
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("failed to get %s (%d)", url, response.StatusCode)
	}

	var responseRaw apiResponse
	err = json.NewDecoder(response.Body).Decode(&responseRaw)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	payload, err := responseRaw.decodePayload()
	if err != nil {
		return "", fmt.Errorf("failed to decode payload: %w", err)
	}

	return payload, nil
}
