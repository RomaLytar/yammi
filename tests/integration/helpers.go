package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// --- Response types ---

type AuthResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ProfileResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// --- API Client ---

type APIClient struct {
	base   string
	client *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		base:   baseURL,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// --- Auth endpoints ---

func (c *APIClient) Register(email, password, name string) (*AuthResponse, int, error) {
	body := map[string]string{"email": email, "password": password, "name": name}
	var resp AuthResponse
	code, err := c.post("/api/v1/auth/register", body, &resp)
	return &resp, code, err
}

func (c *APIClient) Login(email, password string) (*AuthResponse, int, error) {
	body := map[string]string{"email": email, "password": password}
	var resp AuthResponse
	code, err := c.post("/api/v1/auth/login", body, &resp)
	return &resp, code, err
}

func (c *APIClient) RefreshToken(refreshToken string) (*TokenResponse, int, error) {
	body := map[string]string{"refresh_token": refreshToken}
	var resp TokenResponse
	code, err := c.post("/api/v1/auth/refresh", body, &resp)
	return &resp, code, err
}

func (c *APIClient) RevokeToken(refreshToken string) (int, error) {
	body := map[string]string{"refresh_token": refreshToken}
	code, err := c.post("/api/v1/auth/revoke", body, nil)
	return code, err
}

// --- User endpoints ---

func (c *APIClient) GetProfile(userID string) (*ProfileResponse, int, error) {
	var resp ProfileResponse
	code, err := c.get("/api/v1/users/"+userID, &resp)
	return &resp, code, err
}

func (c *APIClient) UpdateProfile(userID, name, avatarURL, bio string) (*ProfileResponse, int, error) {
	body := map[string]string{"name": name, "avatar_url": avatarURL, "bio": bio}
	var resp ProfileResponse
	code, err := c.put("/api/v1/users/"+userID, body, &resp)
	return &resp, code, err
}

func (c *APIClient) DeleteUser(userID string) (int, error) {
	code, err := c.do("DELETE", "/api/v1/users/"+userID, nil, nil)
	return code, err
}

// --- Polling helpers ---

func (c *APIClient) WaitForProfile(t *testing.T, userID string, timeout time.Duration) *ProfileResponse {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, code, err := c.GetProfile(userID)
		if err == nil && code == 200 {
			return resp
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("profile %s not found within %s", userID, timeout)
	return nil
}

func (c *APIClient) WaitForProfileDeletion(t *testing.T, userID string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, code, _ := c.GetProfile(userID)
		if code == 404 {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("profile %s still exists after %s", userID, timeout)
}

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

// --- Assertion helpers ---

func requireStatus(t *testing.T, operation string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: expected HTTP %d, got %d", operation, want, got)
	}
}

func requireNotEmpty(t *testing.T, field, value string) {
	t.Helper()
	if value == "" {
		t.Fatalf("%s must not be empty", field)
	}
}

func requireEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: expected %q, got %q", field, want, got)
	}
}

func requireNotEqual(t *testing.T, field, a, b string) {
	t.Helper()
	if a == b {
		t.Fatalf("%s: expected different values, both are %q", field, a)
	}
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
