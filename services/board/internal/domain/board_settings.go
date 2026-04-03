package domain

import "time"

// BoardSettings — настройки доски (board-scoped).
// Управляет поведением меток: использовать только метки доски или также глобальные пользовательские.
type BoardSettings struct {
	BoardID            string
	UseBoardLabelsOnly bool
	DoneColumnID       *string // колонка "done" для проверки завершения релиза
	SprintDurationDays int     // длительность спринта/релиза в днях (default 14, min 7)
	ReleasesEnabled    bool    // включены ли релизы на доске (default false)
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewBoardSettings создает настройки доски с дефолтными значениями
func NewBoardSettings(boardID string) *BoardSettings {
	now := time.Now()
	return &BoardSettings{
		BoardID:            boardID,
		UseBoardLabelsOnly: false,
		SprintDurationDays: 14,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// Update обновляет настройки доски
func (s *BoardSettings) Update(useBoardLabelsOnly bool, doneColumnID *string, sprintDurationDays int, releasesEnabled bool) {
	s.UseBoardLabelsOnly = useBoardLabelsOnly
	s.DoneColumnID = doneColumnID
	s.SprintDurationDays = sprintDurationDays
	s.ReleasesEnabled = releasesEnabled
	s.UpdatedAt = time.Now()
}
