package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notificationpb "github.com/RomaLytar/yammi/services/notification/api/proto/v1"
	"github.com/RomaLytar/yammi/services/notification/internal/domain"
	"github.com/RomaLytar/yammi/services/notification/internal/usecase"
)

type Handler struct {
	notificationpb.UnimplementedNotificationServiceServer
	listUC     *usecase.ListNotificationsUseCase
	markReadUC *usecase.MarkReadUseCase
	markAllUC  *usecase.MarkAllReadUseCase
	unreadUC   *usecase.GetUnreadCountUseCase
	settingsUC *usecase.SettingsUseCase
}

func NewHandler(
	listUC *usecase.ListNotificationsUseCase,
	markReadUC *usecase.MarkReadUseCase,
	markAllUC *usecase.MarkAllReadUseCase,
	unreadUC *usecase.GetUnreadCountUseCase,
	settingsUC *usecase.SettingsUseCase,
) *Handler {
	return &Handler{
		listUC:     listUC,
		markReadUC: markReadUC,
		markAllUC:  markAllUC,
		unreadUC:   unreadUC,
		settingsUC: settingsUC,
	}
}

func (h *Handler) ListNotifications(ctx context.Context, req *notificationpb.ListNotificationsRequest) (*notificationpb.ListNotificationsResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	notifications, nextCursor, totalUnread, err := h.listUC.Execute(
		ctx, req.GetUserId(), int(req.GetLimit()), req.GetCursor(), req.GetTypeFilter(), req.GetSearch())
	if err != nil {
		return nil, mapDomainError(err)
	}

	pbNotifications := make([]*notificationpb.Notification, 0, len(notifications))
	for _, n := range notifications {
		pbNotifications = append(pbNotifications, toProtoNotification(n))
	}

	return &notificationpb.ListNotificationsResponse{
		Notifications: pbNotifications,
		NextCursor:    nextCursor,
		TotalUnread:   int32(totalUnread),
	}, nil
}

func (h *Handler) MarkAsRead(ctx context.Context, req *notificationpb.MarkAsReadRequest) (*notificationpb.MarkAsReadResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if err := h.markReadUC.Execute(ctx, req.GetUserId(), req.GetNotificationIds()); err != nil {
		return nil, mapDomainError(err)
	}

	return &notificationpb.MarkAsReadResponse{}, nil
}

func (h *Handler) MarkAllAsRead(ctx context.Context, req *notificationpb.MarkAllAsReadRequest) (*notificationpb.MarkAllAsReadResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if err := h.markAllUC.Execute(ctx, req.GetUserId()); err != nil {
		return nil, mapDomainError(err)
	}

	return &notificationpb.MarkAllAsReadResponse{}, nil
}

func (h *Handler) GetUnreadCount(ctx context.Context, req *notificationpb.GetUnreadCountRequest) (*notificationpb.GetUnreadCountResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	count, err := h.unreadUC.Execute(ctx, req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &notificationpb.GetUnreadCountResponse{Count: int32(count)}, nil
}

func (h *Handler) GetSettings(ctx context.Context, req *notificationpb.GetSettingsRequest) (*notificationpb.GetSettingsResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	settings, err := h.settingsUC.Get(ctx, req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &notificationpb.GetSettingsResponse{
		Settings: toProtoSettings(settings),
	}, nil
}

func (h *Handler) UpdateSettings(ctx context.Context, req *notificationpb.UpdateSettingsRequest) (*notificationpb.UpdateSettingsResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	settings, err := h.settingsUC.Update(ctx, req.GetUserId(), req.GetEnabled(), req.GetRealtimeEnabled())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &notificationpb.UpdateSettingsResponse{
		Settings: toProtoSettings(settings),
	}, nil
}

func toProtoNotification(n *domain.Notification) *notificationpb.Notification {
	return &notificationpb.Notification{
		Id:        n.ID,
		UserId:    n.UserID,
		Type:      string(n.Type),
		Title:     n.Title,
		Message:   n.Message,
		Metadata:  n.Metadata,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
	}
}

func toProtoSettings(s *domain.NotificationSettings) *notificationpb.NotificationSettings {
	return &notificationpb.NotificationSettings{
		UserId:          s.UserID,
		Enabled:         s.Enabled,
		RealtimeEnabled: s.RealtimeEnabled,
	}
}

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotificationNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrEmptyUserID):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrEmptyTitle):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrEmptyType):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
