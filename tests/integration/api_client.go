package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// --- API Client ---

type APIClient struct {
	base        string
	client      *http.Client
	accessToken string // Bearer token для авторизованных запросов
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		base:   baseURL,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// SetToken устанавливает access token для последующих запросов.
func (c *APIClient) SetToken(token string) {
	c.accessToken = token
}

// WaitForHealthy ожидает готовности API Gateway.
func (c *APIClient) WaitForHealthy(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := c.client.Get(c.base + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("API Gateway not healthy after %s", timeout)
}

// --- HTTP internals ---

func (c *APIClient) get(path string, out interface{}) (int, error) {
	return c.do("GET", path, nil, out)
}

func (c *APIClient) post(path string, body interface{}, out interface{}) (int, error) {
	return c.do("POST", path, body, out)
}

func (c *APIClient) put(path string, body interface{}, out interface{}) (int, error) {
	return c.do("PUT", path, body, out)
}

func (c *APIClient) do(method, path string, body interface{}, out interface{}) (int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.base+path, reqBody)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if out != nil && len(respBody) > 0 {
		json.Unmarshal(respBody, out)
	}

	return resp.StatusCode, nil
}
