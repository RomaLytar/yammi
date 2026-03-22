package usecase

import (
	"context"
	"log"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type CreateNotificationUseCase struct {
	repo           NotificationRepository
	settings       SettingsRepository
	publisher      EventPublisher
	boardEventRepo BoardEventRepository
	unreadCounter  UnreadCounter
	memberRepo     BoardMemberRepository
}

func NewCreateNotificationUseCase(
	repo NotificationRepository,
	settings SettingsRepository,
	publisher EventPublisher,
	boardEventRepo BoardEventRepository,
	unreadCounter UnreadCounter,
	memberRepo BoardMemberRepository,
) *CreateNotificationUseCase {
	return &CreateNotificationUseCase{
		repo:           repo,
		settings:       settings,
		publisher:      publisher,
		boardEventRepo: boardEventRepo,
		unreadCounter:  unreadCounter,
		memberRepo:     memberRepo,
	}
}

func (uc *CreateNotificationUseCase) Execute(ctx context.Context, userID string, ntype domain.NotificationType, title, message string, metadata map[string]string) error {
	// Проверяем настройки пользователя
	s, err := uc.settings.Get(ctx, userID)
	if err != nil {
		log.Printf("failed to get settings for user %s, using defaults: %v", userID, err)
		s = domain.DefaultSettings(userID)
	}

	if !s.Enabled {
		return nil
	}

	n, err := domain.NewNotification(userID, ntype, title, message, metadata)
	if err != nil {
		return err
	}

	if err := uc.repo.Create(ctx, n); err != nil {
		return err
	}

	// Инкрементируем Redis-счётчик непрочитанных
	if uc.unreadCounter != nil {
		if err := uc.unreadCounter.Increment(ctx, userID); err != nil {
			log.Printf("failed to increment unread counter for user %s: %v", userID, err)
		}
	}

	// Публикуем событие для WebSocket доставки (счётчик + toast)
	if uc.publisher != nil {
		go func() {
			if err := uc.publisher.PublishNotificationCreated(context.Background(), n); err != nil {
				log.Printf("failed to publish notification.created for user %s: %v", userID, err)
			}
		}()
	}

	return nil
}

// NotificationRequest — запрос на создание уведомления (для batch-операций).
type NotificationRequest struct {
	UserID   string
	Type     domain.NotificationType
	Title    string
	Message  string
	Metadata map[string]string
}

// BatchExecute создаёт уведомления для списка пользователей одним batch INSERT.
// Возвращает количество реально созданных уведомлений.
func (uc *CreateNotificationUseCase) BatchExecute(ctx context.Context, requests []NotificationRequest) (int, error) {
	if len(requests) == 0 {
		return 0, nil
	}

	// 1. Собрать уникальные userIDs
	userIDs := make([]string, 0, len(requests))
	seen := make(map[string]bool, len(requests))
	for _, r := range requests {
		if !seen[r.UserID] {
			seen[r.UserID] = true
			userIDs = append(userIDs, r.UserID)
		}
	}

	// 2. Batch-fetch настроек (1 запрос или кеш)
	settingsMap, err := uc.settings.BatchGet(ctx, userIDs)
	if err != nil {
		log.Printf("failed to batch get settings, using defaults: %v", err)
		settingsMap = make(map[string]*domain.NotificationSettings, len(userIDs))
		for _, uid := range userIDs {
			settingsMap[uid] = domain.DefaultSettings(uid)
		}
	}

	// 3. Построить уведомления, отфильтровать disabled
	var notifications []*domain.Notification
	for _, r := range requests {
		s := settingsMap[r.UserID]
		if s == nil {
			s = domain.DefaultSettings(r.UserID)
		}
		if !s.Enabled {
			continue
		}

		n, err := domain.NewNotification(r.UserID, r.Type, r.Title, r.Message, r.Metadata)
		if err != nil {
			log.Printf("failed to create notification for user %s: %v", r.UserID, err)
			continue
		}
		notifications = append(notifications, n)
	}

	if len(notifications) == 0 {
		return 0, nil
	}

	// 4. Batch INSERT (1 запрос)
	if err := uc.repo.BatchCreate(ctx, notifications); err != nil {
		return 0, err
	}

	// 5. Batch publish в NATS (async)
	if uc.publisher != nil {
		go func() {
			if err := uc.publisher.PublishNotificationsBatch(context.Background(), notifications); err != nil {
				log.Printf("failed to batch publish notifications: %v", err)
			}
		}()
	}

	return len(notifications), nil
}

// CreateBoardEvent создаёт один board event и инкрементирует счётчики для всех участников доски.
// Заменяет fan-out (1 event → N INSERT) на event-sourcing (1 event → 1 INSERT + N Redis INCR).
func (uc *CreateNotificationUseCase) CreateBoardEvent(ctx context.Context, boardID, actorID string, eventType domain.NotificationType, title, message string, metadata map[string]string) error {
	// 1. Создаём board event
	event := domain.NewBoardEvent(boardID, actorID, eventType, title, message, metadata)

	// 2. Сохраняем в board_events (1 INSERT)
	if err := uc.boardEventRepo.Create(ctx, event); err != nil {
		return err
	}

	// 3. Получаем участников доски
	memberIDs, err := uc.memberRepo.ListMemberIDs(ctx, boardID)
	if err != nil {
		log.Printf("failed to list members for board %s: %v", boardID, err)
		return nil // event сохранён, ошибка не критична
	}

	// 4. Фильтруем актора
	var recipientIDs []string
	for _, id := range memberIDs {
		if id == actorID {
			continue
		}
		recipientIDs = append(recipientIDs, id)
	}

	if len(recipientIDs) == 0 {
		return nil
	}

	// 5. Проверяем настройки (batch)
	settingsMap, err := uc.settings.BatchGet(ctx, recipientIDs)
	if err != nil {
		log.Printf("failed to batch get settings, using defaults: %v", err)
		settingsMap = make(map[string]*domain.NotificationSettings)
	}

	var enabledIDs []string
	for _, uid := range recipientIDs {
		s := settingsMap[uid]
		if s == nil {
			s = domain.DefaultSettings(uid)
		}
		if s.Enabled {
			enabledIDs = append(enabledIDs, uid)
		}
	}

	if len(enabledIDs) == 0 {
		return nil
	}

	// 6. Инкрементируем Redis-счётчики (pipeline)
	if uc.unreadCounter != nil {
		if err := uc.unreadCounter.IncrementMany(ctx, enabledIDs); err != nil {
			log.Printf("failed to increment unread counters for board %s: %v", boardID, err)
		}
	}

	// 7. Публикуем ОДНО событие для WebSocket push (gateway разошлёт подписчикам board)
	if uc.publisher != nil {
		go func() {
			if err := uc.publisher.PublishBoardEventNotification(context.Background(), event); err != nil {
				log.Printf("failed to publish board event notification for board %s: %v", boardID, err)
			}
		}()
	}

	return nil
}
