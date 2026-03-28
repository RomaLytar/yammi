package domain

import (
	"time"

	"github.com/google/uuid"
)

// BoardColumnTemplateData — данные колонки в шаблоне доски
type BoardColumnTemplateData struct {
	Title    string `json:"title"`
	Position int    `json:"position"`
}

// LabelTemplateData — данные метки в шаблоне доски
type LabelTemplateData struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// BoardTemplate — шаблон доски (user-scoped)
type BoardTemplate struct {
	ID          string
	UserID      string
	Name        string
	Description string
	ColumnsData []BoardColumnTemplateData
	LabelsData  []LabelTemplateData
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewBoardTemplate создает новый шаблон доски с валидацией
func NewBoardTemplate(id, userID, name, description string, columnsData []BoardColumnTemplateData, labelsData []LabelTemplateData) (*BoardTemplate, error) {
	if userID == "" {
		return nil, ErrEmptyOwnerID
	}

	if name == "" {
		return nil, ErrEmptyTemplateName
	}

	if id == "" {
		id = uuid.NewString()
	}

	if columnsData == nil {
		columnsData = []BoardColumnTemplateData{}
	}

	if labelsData == nil {
		labelsData = []LabelTemplateData{}
	}

	now := time.Now()
	return &BoardTemplate{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Description: description,
		ColumnsData: columnsData,
		LabelsData:  labelsData,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
