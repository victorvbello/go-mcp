package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type RequestResult struct {
	StatusCode int
	RawBody    string
}
type RequestHeaders map[string]string

func HttpRequest(method string, url string, headers RequestHeaders, reqBody io.Reader) (RequestResult, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resultBody := RequestResult{}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return resultBody, fmt.Errorf("http.NewRequest %w", err)
	}
	if len(headers) > 0 {
		for key, data := range headers {
			req.Header.Set(key, data)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return resultBody, fmt.Errorf("client.Do %w", err)
	}
	defer resp.Body.Close()

	resultBody.StatusCode = resp.StatusCode

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resultBody, fmt.Errorf("ioutil.ReadAll %w", err)
	}
	if !json.Valid(body) {
		resultBody.RawBody = string(body)
		return resultBody, nil
	}
	err = json.Unmarshal(body, &resultBody)
	if err != nil {
		return resultBody, fmt.Errorf("json.Unmarshal %w", err)
	}
	return resultBody, nil
}
