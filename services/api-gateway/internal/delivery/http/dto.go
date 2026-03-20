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

// --- Board responses ---

type boardResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	Version     int32  `json:"version"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type columnResponse struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	Title     string `json:"title"`
	Position  int32  `json:"position"`
	Version   int32  `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type cardResponse struct {
	ID          string `json:"id"`
	ColumnID    string `json:"column_id"`
	BoardID     string `json:"board_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    string `json:"position"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	Version     int32  `json:"version"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type memberResponse struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	Version  int32  `json:"version"`
	JoinedAt string `json:"joined_at"`
}

// --- Generic responses ---

type statusResponse struct {
	Status string `json:"status"`
}
