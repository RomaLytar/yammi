package integration

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

// --- Unauthenticated auth requests (для тестов 401/403) ---

// RefreshTokenNoAuth делает refresh без access token.
func (c *APIClient) RefreshTokenNoAuth(refreshToken string) (*TokenResponse, int, error) {
	saved := c.accessToken
	c.accessToken = ""
	defer func() { c.accessToken = saved }()
	return c.RefreshToken(refreshToken)
}

// RevokeTokenNoAuth делает revoke без access token.
func (c *APIClient) RevokeTokenNoAuth(refreshToken string) (int, error) {
	saved := c.accessToken
	c.accessToken = ""
	defer func() { c.accessToken = saved }()
	return c.RevokeToken(refreshToken)
}
