package integration

import (
	"testing"
	"time"
)

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

// --- Unauthenticated user requests (для тестов 401/403) ---

// GetProfileNoAuth делает GET без токена.
func (c *APIClient) GetProfileNoAuth(userID string) (*ProfileResponse, int, error) {
	saved := c.accessToken
	c.accessToken = ""
	defer func() { c.accessToken = saved }()
	return c.GetProfile(userID)
}

// DeleteUserNoAuth делает DELETE без токена.
func (c *APIClient) DeleteUserNoAuth(userID string) (int, error) {
	saved := c.accessToken
	c.accessToken = ""
	defer func() { c.accessToken = saved }()
	return c.DeleteUser(userID)
}

// UpdateProfileNoAuth делает PUT без токена.
func (c *APIClient) UpdateProfileNoAuth(userID, name, avatarURL, bio string) (*ProfileResponse, int, error) {
	saved := c.accessToken
	c.accessToken = ""
	defer func() { c.accessToken = saved }()
	return c.UpdateProfile(userID, name, avatarURL, bio)
}

// GetProfileAs делает GET с указанным токеном (для тестов чужого доступа).
func (c *APIClient) GetProfileAs(userID, token string) (*ProfileResponse, int, error) {
	saved := c.accessToken
	c.accessToken = token
	defer func() { c.accessToken = saved }()
	return c.GetProfile(userID)
}

// DeleteUserAs делает DELETE с указанным токеном.
func (c *APIClient) DeleteUserAs(userID, token string) (int, error) {
	saved := c.accessToken
	c.accessToken = token
	defer func() { c.accessToken = saved }()
	return c.DeleteUser(userID)
}

// UpdateProfileAs делает PUT с указанным токеном.
func (c *APIClient) UpdateProfileAs(userID, token, name, avatarURL, bio string) (*ProfileResponse, int, error) {
	saved := c.accessToken
	c.accessToken = token
	defer func() { c.accessToken = saved }()
	return c.UpdateProfile(userID, name, avatarURL, bio)
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
