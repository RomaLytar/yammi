package http

import (
	"encoding/json"
	"net/http"

	notificationpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/notification"
)

type NotificationHandler struct {
	client notificationpb.NotificationServiceClient
}

func NewNotificationHandler(client notificationpb.NotificationServiceClient) *NotificationHandler {
	return &NotificationHandler{client: client}
}

// ListNotifications GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseIntQueryParam(r, "limit", 20)
	cursor := r.URL.Query().Get("cursor")
	typeFilter := r.URL.Query().Get("type")
	search := r.URL.Query().Get("search")

	resp, err := h.client.ListNotifications(r.Context(), &notificationpb.ListNotificationsRequest{
		UserId:     userID,
		Limit:      int32(limit),
		Cursor:     cursor,
		TypeFilter: typeFilter,
		Search:     search,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	notifications := make([]notificationResponse, 0, len(resp.Notifications))
	for _, n := range resp.Notifications {
		notifications = append(notifications, mapNotificationFromProto(n))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"next_cursor":   resp.NextCursor,
		"total_unread":  resp.TotalUnread,
	})
}

// MarkAsRead POST /api/v1/notifications/read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		NotificationIDs []string `json:"notification_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := h.client.MarkAsRead(r.Context(), &notificationpb.MarkAsReadRequest{
		UserId:          userID,
		NotificationIds: req.NotificationIDs,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "ok"})
}

// MarkAllAsRead POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	_, err := h.client.MarkAllAsRead(r.Context(), &notificationpb.MarkAllAsReadRequest{
		UserId: userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "ok"})
}

// GetUnreadCount GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.client.GetUnreadCount(r.Context(), &notificationpb.GetUnreadCountRequest{
		UserId: userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count": resp.Count,
	})
}

// GetSettings GET /api/v1/notifications/settings
func (h *NotificationHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.client.GetSettings(r.Context(), &notificationpb.GetSettingsRequest{
		UserId: userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"settings": settingsResponse{
			Enabled:         resp.Settings.Enabled,
			RealtimeEnabled: resp.Settings.RealtimeEnabled,
		},
	})
}

// UpdateSettings PUT /api/v1/notifications/settings
func (h *NotificationHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Enabled         bool `json:"enabled"`
		RealtimeEnabled bool `json:"realtime_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateSettings(r.Context(), &notificationpb.UpdateSettingsRequest{
		UserId:          userID,
		Enabled:         req.Enabled,
		RealtimeEnabled: req.RealtimeEnabled,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"settings": settingsResponse{
			Enabled:         resp.Settings.Enabled,
			RealtimeEnabled: resp.Settings.RealtimeEnabled,
		},
	})
}

func mapNotificationFromProto(n *notificationpb.Notification) notificationResponse {
	return notificationResponse{
		ID:        n.Id,
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Metadata:  n.Metadata,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
	}
}
