package http

import "strconv"

const (
	maxTitleLen       = 500
	maxDescriptionLen = 5000
	maxNameLen        = 255
	maxSearchLen      = 200
	maxColorLen       = 7 // #FFFFFF
	maxContentLen     = 10000
)

func validateStringLen(s, field string, max int) string {
	if len(s) > max {
		return field + " is too long, max " + strconv.Itoa(max) + " characters"
	}
	return ""
}

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
	DueDate     string `json:"due_date,omitempty"`
	Priority    string `json:"priority"`
	TaskType    string `json:"task_type"`
	ReleaseID   string `json:"release_id,omitempty"`
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

// --- Checklist responses ---

type checklistResponse struct {
	ID        string                  `json:"id"`
	CardID    string                  `json:"card_id"`
	BoardID   string                  `json:"board_id"`
	Title     string                  `json:"title"`
	Position  int32                   `json:"position"`
	Items     []checklistItemResponse `json:"items"`
	Progress  int32                   `json:"progress"`
	CreatedAt string                  `json:"created_at"`
	UpdatedAt string                  `json:"updated_at"`
}

type checklistItemResponse struct {
	ID          string `json:"id"`
	ChecklistID string `json:"checklist_id"`
	Title       string `json:"title"`
	IsChecked   bool   `json:"is_checked"`
	Position    int32  `json:"position"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// --- Card Link responses ---

type cardLinkResponse struct {
	ID              string `json:"id"`
	ParentID        string `json:"parent_id"`
	ChildID         string `json:"child_id"`
	BoardID         string `json:"board_id"`
	LinkType        string `json:"link_type"`
	ChildTitle      string `json:"child_title,omitempty"`
	ChildColumnName string `json:"child_column_name,omitempty"`
	CreatedAt       string `json:"created_at"`
}

// --- Custom Field responses ---

type customFieldDefResponse struct {
	ID        string   `json:"id"`
	BoardID   string   `json:"board_id"`
	Name      string   `json:"name"`
	FieldType string   `json:"field_type"`
	Options   []string `json:"options,omitempty"`
	Position  int32    `json:"position"`
	Required  bool     `json:"required"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type customFieldValueResponse struct {
	ID          string   `json:"id"`
	CardID      string   `json:"card_id"`
	BoardID     string   `json:"board_id"`
	FieldID     string   `json:"field_id"`
	ValueText   *string  `json:"value_text,omitempty"`
	ValueNumber *float64 `json:"value_number,omitempty"`
	ValueDate   *string  `json:"value_date,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// --- Automation Rule responses ---

type automationRuleResponse struct {
	ID            string            `json:"id"`
	BoardID       string            `json:"board_id"`
	Name          string            `json:"name"`
	Enabled       bool              `json:"enabled"`
	TriggerType   string            `json:"trigger_type"`
	TriggerConfig map[string]string `json:"trigger_config"`
	ActionType    string            `json:"action_type"`
	ActionConfig  map[string]string `json:"action_config"`
	CreatedBy     string            `json:"created_by"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

type automationExecutionResponse struct {
	ID             string `json:"id"`
	RuleID         string `json:"rule_id"`
	BoardID        string `json:"board_id"`
	CardID         string `json:"card_id,omitempty"`
	TriggerEventID string `json:"trigger_event_id,omitempty"`
	Status         string `json:"status"`
	ErrorMessage   string `json:"error_message,omitempty"`
	ExecutedAt     string `json:"executed_at"`
}

// --- Board Settings responses ---

type boardSettingsResponse struct {
	BoardID            string `json:"board_id"`
	UseBoardLabelsOnly bool   `json:"use_board_labels_only"`
	DoneColumnID       string `json:"done_column_id,omitempty"`
	SprintDurationDays int32  `json:"sprint_duration_days"`
	ReleasesEnabled    bool   `json:"releases_enabled"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

// --- User Label responses ---

type userLabelResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}

// --- Generic responses ---

// Template DTOs
type boardColumnTemplateDataResponse struct {
	Title    string `json:"title"`
	Position int32  `json:"position"`
}

type labelTemplateDataResponse struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type boardTemplateResponse struct {
	ID          string                            `json:"id"`
	UserID      string                            `json:"user_id"`
	Name        string                            `json:"name"`
	Description string                            `json:"description"`
	ColumnsData []boardColumnTemplateDataResponse `json:"columns_data"`
	LabelsData  []labelTemplateDataResponse       `json:"labels_data"`
	CreatedAt   string                            `json:"created_at"`
	UpdatedAt   string                            `json:"updated_at"`
}

// --- Release responses ---

type releaseResponse struct {
	ID          string `json:"id"`
	BoardID     string `json:"board_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	CreatedBy   string `json:"created_by"`
	Version     int32  `json:"version"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type statusResponse struct {
	Status string `json:"status"`
}
