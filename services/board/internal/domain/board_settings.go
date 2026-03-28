package domain

import "time"

// BoardSettings — настройки доски (board-scoped).
// Управляет поведением меток: использовать только метки доски или также глобальные пользовательские.
type BoardSettings struct {
	BoardID            string
	UseBoardLabelsOnly bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewBoardSettings создает настройки доски с дефолтными значениями
func NewBoardSettings(boardID string) *BoardSettings {
	now := time.Now()
	return &BoardSettings{
		BoardID:            boardID,
		UseBoardLabelsOnly: false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// Update обновляет настройки доски
func (s *BoardSettings) Update(useBoardLabelsOnly bool) {
	s.UseBoardLabelsOnly = useBoardLabelsOnly
	s.UpdatedAt = time.Now()
}
