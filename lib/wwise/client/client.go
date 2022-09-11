package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
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
		return errors.Wrap(err, "failed to marshal auth json")
	}

	response, err := http.Post("https://www.audiokinetic.com/api/login", "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		return errors.Wrap(err, "failed to post auth request")
	}

	if response.StatusCode != 200 {
		return errors.New("failed to authenticate")
	}

	var responseJson struct {
		Code          int    `json:"code"`
		Jwt           string `json:"jwt"`
		SpecialAction bool   `json:"specialAction"`
		Random        string `json:"random"`
	}
	err = json.NewDecoder(response.Body).Decode(&responseJson)
	if err != nil {
		return errors.Wrap(err, "failed to decode auth response")
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
		return "", errors.Wrap(err, "failed to decode payload")
	}
	return string(payloadBytes), nil
}

func (client *WwiseClient) SendRequest(method string, url string, body interface{}) (string, error) {
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal request body")
	}

	if !strings.HasPrefix(url, "https://blob-api.gowwise.com") {
		url = "https://blob-api.gowwise.com" + url
	}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return "", errors.Wrap(err, "failed to create request")
	}

	request.Header.Set("Authorization", "Bearer "+client.auth)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", errors.Wrap(err, "failed to send request")
	}

	if err != nil {
		return "", errors.Wrap(err, "failed to get manifest")
	}

	if response.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("failed to get %s (%d)", url, response.StatusCode))
	}

	var responseRaw apiResponse
	err = json.NewDecoder(response.Body).Decode(&responseRaw)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response")
	}

	payload, err := responseRaw.decodePayload()
	if err != nil {
		return "", errors.Wrap(err, "failed to decode payload")
	}

	return payload, nil
}
