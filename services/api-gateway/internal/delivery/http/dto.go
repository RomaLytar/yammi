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
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	OwnerID        string `json:"owner_id"`
	Version        int32  `json:"version"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	OwnerName      string `json:"owner_name,omitempty"`
	OwnerAvatarURL string `json:"owner_avatar_url,omitempty"`
}

type columnResponse struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	Title     string `json:"title"`
	Position  int32  `json:"position"`
	Version   int32  `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	CardCount int32  `json:"card_count"`
}

type cardResponse struct {
	ID          string `json:"id"`
	ColumnID    string `json:"column_id"`
	BoardID     string `json:"board_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    string `json:"position"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	CreatorID   string `json:"creator_id"`
	Version     int32  `json:"version"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type memberResponse struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	Version   int32  `json:"version"`
	JoinedAt  string `json:"joined_at"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// --- Attachment responses ---

type attachmentResponse struct {
	ID         string `json:"id"`
	CardID     string `json:"card_id"`
	BoardID    string `json:"board_id"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	MimeType   string `json:"mime_type"`
	UploaderID string `json:"uploader_id"`
	CreatedAt  string `json:"created_at"`
}

// --- Notification responses ---

type notificationResponse struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata"`
	IsRead    bool              `json:"is_read"`
	CreatedAt string            `json:"created_at"`
}

type settingsResponse struct {
	Enabled         bool `json:"enabled"`
	RealtimeEnabled bool `json:"realtime_enabled"`
}

// --- Comment responses ---

type commentResponse struct {
	ID         string `json:"id"`
	CardID     string `json:"card_id"`
	BoardID    string `json:"board_id"`
	AuthorID   string `json:"author_id"`
	ParentID   string `json:"parent_id,omitempty"`
	Content    string `json:"content"`
	ReplyCount int32  `json:"reply_count"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// --- Label responses ---

type labelResponse struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}

// --- Generic responses ---

type statusResponse struct {
	Status string `json:"status"`
}
