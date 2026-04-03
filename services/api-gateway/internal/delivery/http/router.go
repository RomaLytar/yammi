package http

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RomaLytar/yammi/services/api-gateway/internal/infrastructure"
)

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// NewRouter создаёт HTTP роутер. Возвращает handler и функцию shutdown для остановки rate limiter горутин.
func NewRouter(clients *infrastructure.GRPCClients, verifier *infrastructure.JWTVerifier) (http.Handler, func()) {
	mux := http.NewServeMux()
	requireAuth := AuthMiddleware(verifier)

	// Rate limiters — лимиты настраиваются через env, дефолт 50 req/min
	registerLimiter := NewRateLimiter(envInt("RATE_LIMIT_REGISTER", 50), time.Minute)
	loginLimiter := NewRateLimiter(envInt("RATE_LIMIT_LOGIN", 50), time.Minute)
	refreshLimiter := NewRateLimiter(envInt("RATE_LIMIT_REFRESH", 50), time.Minute)
	defaultLimiter := NewRateLimiter(envInt("RATE_LIMIT_DEFAULT", 50), time.Minute)
	rateLimit := RateLimitMiddleware(defaultLimiter)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth routes — публичные (с жёсткими лимитами)
	auth := NewAuthHandler(clients.AuthClient)
	mux.HandleFunc("POST /api/v1/auth/register", RateLimitHandlerFunc(registerLimiter, auth.Register))
	mux.HandleFunc("POST /api/v1/auth/login", RateLimitHandlerFunc(loginLimiter, auth.Login))
	mux.HandleFunc("GET /api/v1/auth/public-key", auth.GetPublicKey)

	// Auth routes — refresh публичный (токен проверяется в Auth Service), revoke требует авторизацию
	mux.HandleFunc("POST /api/v1/auth/refresh", RateLimitHandlerFunc(refreshLimiter, auth.RefreshToken))
	mux.Handle("POST /api/v1/auth/revoke", rateLimit(requireAuth(http.HandlerFunc(auth.RevokeToken))))

	// User routes
	user := NewUserHandler(clients.UserClient, clients.AuthClient)
	mux.Handle("GET /api/v1/users/{id}", rateLimit(requireAuth(http.HandlerFunc(user.GetProfile))))
	mux.Handle("PUT /api/v1/users/{id}", rateLimit(requireAuth(http.HandlerFunc(OwnerOnly(user.UpdateProfile)))))
	mux.Handle("DELETE /api/v1/users/{id}", rateLimit(requireAuth(http.HandlerFunc(OwnerOnly(user.DeleteUser)))))
	mux.Handle("GET /api/v1/users/search", rateLimit(requireAuth(http.HandlerFunc(user.SearchByEmail))))

	// Board routes — все требуют auth
	board := NewBoardHandler(clients.BoardClient, clients.UserClient)
	mux.Handle("POST /api/v1/boards", rateLimit(requireAuth(http.HandlerFunc(board.CreateBoard))))
	mux.Handle("GET /api/v1/boards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.GetBoard))))
	mux.Handle("GET /api/v1/boards", rateLimit(requireAuth(http.HandlerFunc(board.ListBoards))))
	mux.Handle("PUT /api/v1/boards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateBoard))))
	mux.Handle("POST /api/v1/boards/delete", rateLimit(requireAuth(http.HandlerFunc(board.DeleteBoards))))

	// Column routes
	mux.Handle("POST /api/v1/boards/{id}/columns", rateLimit(requireAuth(http.HandlerFunc(board.AddColumn))))
	mux.Handle("GET /api/v1/boards/{id}/columns", rateLimit(requireAuth(http.HandlerFunc(board.GetColumns))))
	mux.Handle("PUT /api/v1/boards/{id}/columns/reorder", rateLimit(requireAuth(http.HandlerFunc(board.ReorderColumns))))
	mux.Handle("PUT /api/v1/columns/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateColumn))))
	mux.Handle("DELETE /api/v1/columns/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteColumn))))

	// Card routes
	mux.Handle("POST /api/v1/columns/{id}/cards", rateLimit(requireAuth(http.HandlerFunc(board.CreateCard))))
	mux.Handle("GET /api/v1/cards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.GetCard))))
	mux.Handle("GET /api/v1/columns/{id}/cards", rateLimit(requireAuth(http.HandlerFunc(board.GetCards))))
	mux.Handle("PUT /api/v1/cards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateCard))))
	mux.Handle("PUT /api/v1/cards/{id}/move", rateLimit(requireAuth(http.HandlerFunc(board.MoveCard))))
	mux.Handle("POST /api/v1/cards/delete", rateLimit(requireAuth(http.HandlerFunc(board.DeleteCards))))
	mux.Handle("PUT /api/v1/cards/{id}/assign", rateLimit(requireAuth(http.HandlerFunc(board.AssignCard))))
	mux.Handle("DELETE /api/v1/cards/{id}/assign", rateLimit(requireAuth(http.HandlerFunc(board.UnassignCard))))
	mux.Handle("GET /api/v1/cards/{id}/activity", rateLimit(requireAuth(http.HandlerFunc(board.GetCardActivity))))
	mux.Handle("GET /api/v1/boards/{id}/cards/search", rateLimit(requireAuth(http.HandlerFunc(board.SearchBoardCards))))

	// Member routes
	mux.Handle("POST /api/v1/boards/{id}/members", rateLimit(requireAuth(http.HandlerFunc(board.AddMember))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/members/{userId}", rateLimit(requireAuth(http.HandlerFunc(board.RemoveMember))))
	mux.Handle("GET /api/v1/boards/{id}/members", rateLimit(requireAuth(http.HandlerFunc(board.ListMembers))))

	// Attachment routes
	mux.Handle("POST /api/v1/cards/{id}/attachments/upload-url", rateLimit(requireAuth(http.HandlerFunc(board.CreateUploadURL))))
	mux.Handle("POST /api/v1/attachments/{id}/confirm", rateLimit(requireAuth(http.HandlerFunc(board.ConfirmUpload))))
	mux.Handle("GET /api/v1/attachments/{id}/download-url", rateLimit(requireAuth(http.HandlerFunc(board.GetDownloadURL))))
	mux.Handle("GET /api/v1/cards/{id}/attachments", rateLimit(requireAuth(http.HandlerFunc(board.ListAttachments))))
	mux.Handle("DELETE /api/v1/attachments/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteAttachment))))

	// Label routes
	mux.Handle("POST /api/v1/boards/{id}/labels", rateLimit(requireAuth(http.HandlerFunc(board.CreateLabel))))
	mux.Handle("GET /api/v1/boards/{id}/labels", rateLimit(requireAuth(http.HandlerFunc(board.ListLabels))))
	mux.Handle("PUT /api/v1/boards/{boardId}/labels/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateLabel))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/labels/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteLabel))))
	mux.Handle("POST /api/v1/boards/{boardId}/cards/{cardId}/labels", rateLimit(requireAuth(http.HandlerFunc(board.AddLabelToCard))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/cards/{cardId}/labels/{labelId}", rateLimit(requireAuth(http.HandlerFunc(board.RemoveLabelFromCard))))
	mux.Handle("GET /api/v1/boards/{boardId}/cards/{cardId}/labels", rateLimit(requireAuth(http.HandlerFunc(board.GetCardLabels))))

	// Checklist routes
	mux.Handle("POST /api/v1/boards/{boardId}/cards/{cardId}/checklists", rateLimit(requireAuth(http.HandlerFunc(board.CreateChecklist))))
	mux.Handle("GET /api/v1/boards/{boardId}/cards/{cardId}/checklists", rateLimit(requireAuth(http.HandlerFunc(board.GetChecklists))))
	mux.Handle("PUT /api/v1/boards/{boardId}/checklists/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateChecklist))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/checklists/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteChecklist))))
	mux.Handle("POST /api/v1/boards/{boardId}/checklists/{checklistId}/items", rateLimit(requireAuth(http.HandlerFunc(board.CreateChecklistItem))))
	mux.Handle("PUT /api/v1/boards/{boardId}/checklist-items/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateChecklistItem))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/checklist-items/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteChecklistItem))))
	mux.Handle("PUT /api/v1/boards/{boardId}/checklist-items/{id}/toggle", rateLimit(requireAuth(http.HandlerFunc(board.ToggleChecklistItem))))

	// Card Link routes
	mux.Handle("POST /api/v1/boards/{boardId}/cards/{cardId}/links", rateLimit(requireAuth(http.HandlerFunc(board.LinkCards))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/card-links/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UnlinkCards))))
	mux.Handle("GET /api/v1/boards/{boardId}/cards/{cardId}/children", rateLimit(requireAuth(http.HandlerFunc(board.GetCardChildren))))
	mux.Handle("GET /api/v1/boards/{boardId}/cards/{cardId}/parents", rateLimit(requireAuth(http.HandlerFunc(board.GetCardParents))))

	// Custom Field routes
	mux.Handle("POST /api/v1/boards/{id}/custom-fields", rateLimit(requireAuth(http.HandlerFunc(board.CreateCustomField))))
	mux.Handle("GET /api/v1/boards/{id}/custom-fields", rateLimit(requireAuth(http.HandlerFunc(board.ListCustomFields))))
	mux.Handle("PUT /api/v1/boards/{boardId}/custom-fields/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateCustomField))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/custom-fields/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteCustomField))))
	mux.Handle("PUT /api/v1/boards/{boardId}/cards/{cardId}/custom-fields/{fieldId}", rateLimit(requireAuth(http.HandlerFunc(board.SetCustomFieldValue))))
	mux.Handle("GET /api/v1/boards/{boardId}/cards/{cardId}/custom-fields", rateLimit(requireAuth(http.HandlerFunc(board.GetCardCustomFields))))

	// Automation Rule routes
	mux.Handle("POST /api/v1/boards/{id}/automations", rateLimit(requireAuth(http.HandlerFunc(board.CreateAutomationRule))))
	mux.Handle("GET /api/v1/boards/{id}/automations", rateLimit(requireAuth(http.HandlerFunc(board.ListAutomationRules))))
	mux.Handle("PUT /api/v1/boards/{boardId}/automations/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateAutomationRule))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/automations/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteAutomationRule))))
	mux.Handle("GET /api/v1/boards/{boardId}/automations/{id}/history", rateLimit(requireAuth(http.HandlerFunc(board.GetAutomationHistory))))

	// Board Settings routes
	mux.Handle("GET /api/v1/boards/{id}/settings", rateLimit(requireAuth(http.HandlerFunc(board.GetBoardSettings))))
	mux.Handle("PUT /api/v1/boards/{id}/settings", rateLimit(requireAuth(http.HandlerFunc(board.UpdateBoardSettings))))

	// User Label routes
	mux.Handle("POST /api/v1/user-labels", rateLimit(requireAuth(http.HandlerFunc(board.CreateUserLabel))))
	mux.Handle("GET /api/v1/user-labels", rateLimit(requireAuth(http.HandlerFunc(board.ListUserLabels))))
	mux.Handle("PUT /api/v1/user-labels/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateUserLabel))))
	mux.Handle("DELETE /api/v1/user-labels/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteUserLabel))))

	// Available Labels route
	mux.Handle("GET /api/v1/boards/{id}/available-labels", rateLimit(requireAuth(http.HandlerFunc(board.GetAvailableLabels))))

	// Release routes
	mux.Handle("POST /api/v1/boards/{id}/releases", rateLimit(requireAuth(http.HandlerFunc(board.CreateRelease))))
	mux.Handle("GET /api/v1/boards/{id}/releases", rateLimit(requireAuth(http.HandlerFunc(board.ListReleases))))
	mux.Handle("GET /api/v1/boards/{id}/releases/active", rateLimit(requireAuth(http.HandlerFunc(board.GetActiveRelease))))
	mux.Handle("GET /api/v1/boards/{boardId}/releases/{releaseId}", rateLimit(requireAuth(http.HandlerFunc(board.GetRelease))))
	mux.Handle("PUT /api/v1/boards/{boardId}/releases/{releaseId}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateRelease))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/releases/{releaseId}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteRelease))))
	mux.Handle("POST /api/v1/boards/{boardId}/releases/{releaseId}/start", rateLimit(requireAuth(http.HandlerFunc(board.StartRelease))))
	mux.Handle("POST /api/v1/boards/{boardId}/releases/{releaseId}/complete", rateLimit(requireAuth(http.HandlerFunc(board.CompleteRelease))))
	mux.Handle("GET /api/v1/boards/{boardId}/releases/{releaseId}/cards", rateLimit(requireAuth(http.HandlerFunc(board.GetReleaseCards))))
	mux.Handle("POST /api/v1/boards/{boardId}/releases/{releaseId}/cards", rateLimit(requireAuth(http.HandlerFunc(board.AssignCardToRelease))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/releases/{releaseId}/cards/{cardId}", rateLimit(requireAuth(http.HandlerFunc(board.RemoveCardFromRelease))))
	mux.Handle("GET /api/v1/boards/{id}/backlog", rateLimit(requireAuth(http.HandlerFunc(board.GetBacklog))))

	// Board Template routes
	mux.Handle("POST /api/v1/board-templates", rateLimit(requireAuth(http.HandlerFunc(board.CreateBoardTemplate))))
	mux.Handle("GET /api/v1/board-templates", rateLimit(requireAuth(http.HandlerFunc(board.ListBoardTemplates))))
	mux.Handle("DELETE /api/v1/board-templates/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteBoardTemplate))))
	mux.Handle("POST /api/v1/boards/from-template", rateLimit(requireAuth(http.HandlerFunc(board.CreateBoardFromTemplate))))

	// Comment routes
	comment := NewCommentHandler(clients.CommentClient)
	mux.Handle("POST /api/v1/cards/{id}/comments", rateLimit(requireAuth(http.HandlerFunc(comment.CreateComment))))
	mux.Handle("GET /api/v1/cards/{id}/comments", rateLimit(requireAuth(http.HandlerFunc(comment.ListComments))))
	mux.Handle("PUT /api/v1/comments/{id}", rateLimit(requireAuth(http.HandlerFunc(comment.UpdateComment))))
	mux.Handle("DELETE /api/v1/comments/{id}", rateLimit(requireAuth(http.HandlerFunc(comment.DeleteComment))))
	mux.Handle("GET /api/v1/cards/{id}/comments/count", rateLimit(requireAuth(http.HandlerFunc(comment.GetCommentCount))))

	// Notification routes
	notification := NewNotificationHandler(clients.NotificationClient)
	mux.Handle("GET /api/v1/notifications", rateLimit(requireAuth(http.HandlerFunc(notification.ListNotifications))))
	mux.Handle("POST /api/v1/notifications/read", rateLimit(requireAuth(http.HandlerFunc(notification.MarkAsRead))))
	mux.Handle("POST /api/v1/notifications/read-all", rateLimit(requireAuth(http.HandlerFunc(notification.MarkAllAsRead))))
	mux.Handle("GET /api/v1/notifications/unread-count", rateLimit(requireAuth(http.HandlerFunc(notification.GetUnreadCount))))
	mux.Handle("GET /api/v1/notifications/settings", rateLimit(requireAuth(http.HandlerFunc(notification.GetSettings))))
	mux.Handle("PUT /api/v1/notifications/settings", rateLimit(requireAuth(http.HandlerFunc(notification.UpdateSettings))))

	shutdown := func() {
		registerLimiter.Stop()
		loginLimiter.Stop()
		refreshLimiter.Stop()
		defaultLimiter.Stop()
	}

	// Оборачиваем в security middlewares
	var handler http.Handler = mux
	handler = MaxBodyMiddleware(1 << 20)(handler) // 1 MB
	handler = SecurityHeadersMiddleware(handler)
	handler = CORSMiddleware(handler)

	return handler, shutdown
}
