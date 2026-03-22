package domain

import (
	"errors"
	"testing"
)

func TestNewActivity(t *testing.T) {
	tests := []struct {
		name         string
		cardID       string
		boardID      string
		actorID      string
		activityType ActivityType
		description  string
		changes      map[string]string
		wantErr      error
	}{
		{
			name:         "valid card_created activity",
			cardID:       "card-123",
			boardID:      "board-123",
			actorID:      "user-123",
			activityType: ActivityCardCreated,
			description:  "Карточка \"Task 1\" создана",
			changes:      nil,
			wantErr:      nil,
		},
		{
			name:         "valid card_updated activity with changes",
			cardID:       "card-123",
			boardID:      "board-123",
			actorID:      "user-123",
			activityType: ActivityCardUpdated,
			description:  "Карточка \"Task 1\" обновлена",
			changes:      map[string]string{"old_title": "Old", "new_title": "New"},
			wantErr:      nil,
		},
		{
			name:         "valid card_moved activity",
			cardID:       "card-123",
			boardID:      "board-123",
			actorID:      "user-123",
			activityType: ActivityCardMoved,
			description:  "Карточка перемещена",
			changes:      map[string]string{"from_column_id": "col-1", "to_column_id": "col-2"},
			wantErr:      nil,
		},
		{
			name:         "valid card_assigned activity",
			cardID:       "card-123",
			boardID:      "board-123",
			actorID:      "user-123",
			activityType: ActivityCardAssigned,
			description:  "Исполнитель назначен",
			changes:      map[string]string{"assignee_id": "user-456"},
			wantErr:      nil,
		},
		{
			name:         "valid card_unassigned activity",
			cardID:       "card-123",
			boardID:      "board-123",
			actorID:      "user-123",
			activityType: ActivityCardUnassigned,
			description:  "Исполнитель снят",
			changes:      map[string]string{"prev_assignee_id": "user-456"},
			wantErr:      nil,
		},
		{
			name:         "valid card_deleted activity",
			cardID:       "card-123",
			boardID:      "board-123",
			actorID:      "user-123",
			activityType: ActivityCardDeleted,
			description:  "Карточка удалена",
			changes:      nil,
			wantErr:      nil,
		},
		{
			name:         "empty card_id",
			cardID:       "",
			boardID:      "board-123",
			actorID:      "user-123",
			activityType: ActivityCardCreated,
			description:  "Карточка создана",
			changes:      nil,
			wantErr:      ErrCardNotFound,
		},
		{
			name:         "empty board_id",
			cardID:       "card-123",
			boardID:      "",
			actorID:      "user-123",
			activityType: ActivityCardCreated,
			description:  "Карточка создана",
			changes:      nil,
			wantErr:      ErrBoardNotFound,
		},
		{
			name:         "empty actor_id",
			cardID:       "card-123",
			boardID:      "board-123",
			actorID:      "",
			activityType: ActivityCardCreated,
			description:  "Карточка создана",
			changes:      nil,
			wantErr:      ErrEmptyActorID,
		},
		{
			name:         "empty activity_type",
			cardID:       "card-123",
			boardID:      "board-123",
			actorID:      "user-123",
			activityType: "",
			description:  "Карточка создана",
			changes:      nil,
			wantErr:      ErrInvalidActivityType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity, err := NewActivity(tt.cardID, tt.boardID, tt.actorID, tt.activityType, tt.description, tt.changes)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewActivity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				if activity != nil {
					t.Errorf("NewActivity() returned activity when error expected")
				}
				return
			}

			// Проверяем корректность созданной активности
			if activity == nil {
				t.Fatal("NewActivity() returned nil activity")
			}

			if activity.ID == "" {
				t.Error("NewActivity() ID is empty")
			}

			if activity.CardID != tt.cardID {
				t.Errorf("NewActivity() CardID = %v, want %v", activity.CardID, tt.cardID)
			}

			if activity.BoardID != tt.boardID {
				t.Errorf("NewActivity() BoardID = %v, want %v", activity.BoardID, tt.boardID)
			}

			if activity.ActorID != tt.actorID {
				t.Errorf("NewActivity() ActorID = %v, want %v", activity.ActorID, tt.actorID)
			}

			if activity.Type != tt.activityType {
				t.Errorf("NewActivity() Type = %v, want %v", activity.Type, tt.activityType)
			}

			if activity.Description != tt.description {
				t.Errorf("NewActivity() Description = %v, want %v", activity.Description, tt.description)
			}

			if activity.CreatedAt.IsZero() {
				t.Error("NewActivity() CreatedAt is zero")
			}

			// Changes всегда должен быть non-nil (даже если передан nil)
			if activity.Changes == nil {
				t.Error("NewActivity() Changes is nil, want non-nil map")
			}
		})
	}
}
