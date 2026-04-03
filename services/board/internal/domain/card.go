package domain

import (
	"time"

	"github.com/google/uuid"
)

// Priority представляет приоритет карточки
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// IsValid проверяет валидность приоритета
func (p Priority) IsValid() bool {
	return p == PriorityLow || p == PriorityMedium || p == PriorityHigh || p == PriorityCritical
}

// String реализует fmt.Stringer
func (p Priority) String() string {
	return string(p)
}

// TaskType представляет тип задачи
type TaskType string

const (
	TaskTypeBug         TaskType = "bug"
	TaskTypeFeature     TaskType = "feature"
	TaskTypeTask        TaskType = "task"
	TaskTypeImprovement TaskType = "improvement"
)

// IsValid проверяет валидность типа задачи
func (t TaskType) IsValid() bool {
	return t == TaskTypeBug || t == TaskTypeFeature || t == TaskTypeTask || t == TaskTypeImprovement
}

// String реализует fmt.Stringer
func (t TaskType) String() string {
	return string(t)
}

// Card — aggregate root для карточки.
// Отделен от Board/Column для производительности и granular cache.
// Position использует lexorank (string) вместо INT для efficient reordering.
type Card struct {
	ID          string
	ColumnID    string
	Title       string
	Description string
	Position    string     // lexorank (a, am, b, c, ...) — НЕ INT!
	AssigneeID  *string    // опциональный исполнитель
	CreatorID   string     // кто создал карточку
	ReleaseID   *string    // опциональный релиз (nil = бэклог)
	DueDate     *time.Time // опциональный дедлайн
	Priority    Priority   // приоритет (low, medium, high, critical)
	TaskType    TaskType   // тип задачи (bug, feature, task, improvement)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewCard создает новую карточку с валидацией
func NewCard(columnID, title, description, position string, assigneeID *string, creatorID string, dueDate *time.Time, priority Priority, taskType TaskType) (*Card, error) {
	if columnID == "" {
		return nil, ErrColumnNotFound
	}

	if title == "" {
		return nil, ErrEmptyCardTitle
	}

	if err := ValidateLexorank(position); err != nil {
		return nil, err
	}

	// Дефолты для priority и taskType
	if priority == "" {
		priority = PriorityMedium
	}
	if taskType == "" {
		taskType = TaskTypeTask
	}

	if !priority.IsValid() {
		return nil, ErrInvalidPriority
	}
	if !taskType.IsValid() {
		return nil, ErrInvalidTaskType
	}

	now := time.Now()
	return &Card{
		ID:          uuid.NewString(),
		ColumnID:    columnID,
		Title:       title,
		Description: description,
		Position:    position,
		AssigneeID:  assigneeID,
		CreatorID:   creatorID,
		DueDate:     dueDate,
		Priority:    priority,
		TaskType:    taskType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update обновляет метаданные карточки
func (c *Card) Update(title, description string, assigneeID *string, dueDate *time.Time, priority Priority, taskType TaskType) error {
	if title == "" {
		return ErrEmptyCardTitle
	}

	// Дефолты для priority и taskType
	if priority == "" {
		priority = PriorityMedium
	}
	if taskType == "" {
		taskType = TaskTypeTask
	}

	if !priority.IsValid() {
		return ErrInvalidPriority
	}
	if !taskType.IsValid() {
		return ErrInvalidTaskType
	}

	c.Title = title
	c.Description = description
	c.AssigneeID = assigneeID
	c.DueDate = dueDate
	c.Priority = priority
	c.TaskType = taskType
	c.UpdatedAt = time.Now()

	return nil
}

// Move перемещает карточку в другую колонку с новой позицией
func (c *Card) Move(targetColumnID, newPosition string) error {
	if targetColumnID == "" {
		return ErrColumnNotFound
	}

	if err := ValidateLexorank(newPosition); err != nil {
		return err
	}

	c.ColumnID = targetColumnID
	c.Position = newPosition
	c.UpdatedAt = time.Now()

	return nil
}

// Reorder изменяет позицию карточки в текущей колонке
func (c *Card) Reorder(newPosition string) error {
	if err := ValidateLexorank(newPosition); err != nil {
		return err
	}

	c.Position = newPosition
	c.UpdatedAt = time.Now()

	return nil
}
