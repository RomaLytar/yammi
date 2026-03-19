package http

// --- Auth responses ---

type authResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type publicKeyResponse struct {
	PublicKeyPEM string `json:"public_key_pem"`
	Algorithm    string `json:"algorithm"`
}

// --- User responses ---

type profileResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// --- Generic responses ---

type statusResponse struct {
	Status string `json:"status"`
}
