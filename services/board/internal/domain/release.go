package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReleaseStatus представляет состояние жизненного цикла релиза
type ReleaseStatus string

const (
	ReleaseStatusDraft     ReleaseStatus = "draft"
	ReleaseStatusActive    ReleaseStatus = "active"
	ReleaseStatusCompleted ReleaseStatus = "completed"
)

// IsValid проверяет валидность статуса
func (s ReleaseStatus) IsValid() bool {
	return s == ReleaseStatusDraft || s == ReleaseStatusActive || s == ReleaseStatusCompleted
}

func (s ReleaseStatus) String() string {
	return string(s)
}

// MaxReleasesPerBoard — максимальное количество релизов на доску
const MaxReleasesPerBoard = 50

// Release — релиз доски. Жизненный цикл: draft → active → completed.
// Только один релиз может быть active на доске одновременно.
// Completed релизы — read-only (нельзя редактировать, удалять карточки, менять статусы).
type Release struct {
	ID          string
	BoardID     string
	Name        string
	Description string
	Status      ReleaseStatus
	StartDate   *time.Time // планируемая дата начала
	EndDate     *time.Time // планируемая дата окончания
	StartedAt   *time.Time // фактическое время старта (auto)
	CompletedAt *time.Time // фактическое время завершения (auto)
	CreatedBy   string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewRelease создает новый релиз со статусом draft
func NewRelease(boardID, name, description, createdBy string, startDate, endDate *time.Time) (*Release, error) {
	if boardID == "" {
		return nil, ErrBoardNotFound
	}
	if name == "" {
		return nil, ErrEmptyReleaseName
	}
	if createdBy == "" {
		return nil, ErrEmptyActorID
	}

	now := time.Now()
	return &Release{
		ID:          uuid.NewString(),
		BoardID:     boardID,
		Name:        name,
		Description: description,
		Status:      ReleaseStatusDraft,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedBy:   createdBy,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update обновляет имя, описание и даты релиза. Completed релизы нельзя изменять.
func (r *Release) Update(name, description string, startDate, endDate *time.Time) error {
	if r.IsCompleted() {
		return ErrReleaseCompleted
	}
	if name == "" {
		return ErrEmptyReleaseName
	}

	r.Name = name
	r.Description = description
	r.StartDate = startDate
	r.EndDate = endDate
	r.UpdatedAt = time.Now()
	return nil
}

// Start переводит релиз из draft в active. durationDays определяет длительность спринта
// и используется для вычисления EndDate = StartedAt + durationDays.
func (r *Release) Start(durationDays int) error {
	if r.Status != ReleaseStatusDraft {
		return ErrReleaseNotDraft
	}

	now := time.Now()
	endDate := now.AddDate(0, 0, durationDays)
	r.Status = ReleaseStatusActive
	r.StartedAt = &now
	r.StartDate = &now
	r.EndDate = &endDate
	r.UpdatedAt = now
	return nil
}

// Complete переводит релиз из active в completed
func (r *Release) Complete() error {
	if r.Status != ReleaseStatusActive {
		return ErrReleaseNotActive
	}

	now := time.Now()
	r.Status = ReleaseStatusCompleted
	r.CompletedAt = &now
	r.UpdatedAt = now
	return nil
}

// IsDraft проверяет, что релиз в статусе draft
func (r *Release) IsDraft() bool {
	return r.Status == ReleaseStatusDraft
}

// IsActive проверяет, что релиз в статусе active
func (r *Release) IsActive() bool {
	return r.Status == ReleaseStatusActive
}

// IsCompleted проверяет, что релиз завершён
func (r *Release) IsCompleted() bool {
	return r.Status == ReleaseStatusCompleted
}

// IncrementVersion увеличивает версию для оптимистичной блокировки
func (r *Release) IncrementVersion() {
	r.Version++
}
