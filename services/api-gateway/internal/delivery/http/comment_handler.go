package http

import (
	"encoding/json"
	"net/http"

	commentpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/comment"
)

// CommentHandler обрабатывает HTTP запросы для Comment Service
type CommentHandler struct {
	client commentpb.CommentServiceClient
}

// NewCommentHandler создаёт новый обработчик комментариев
func NewCommentHandler(client commentpb.CommentServiceClient) *CommentHandler {
	return &CommentHandler{client: client}
}

// CreateComment POST /api/v1/cards/{id}/comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	var req struct {
		BoardID  string `json:"board_id"`
		Content  string `json:"content"`
		ParentID string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.BoardID == "" {
		writeError(w, http.StatusBadRequest, "board_id is required")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	if msg := validateStringLen(req.Content, "content", maxContentLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	resp, err := h.client.CreateComment(r.Context(), &commentpb.CreateCommentRequest{
		CardId:   cardID,
		BoardId:  req.BoardID,
		UserId:   userID,
		Content:  req.Content,
		ParentId: req.ParentID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"comment": mapCommentFromProto(resp.Comment),
	})
}

// ListComments GET /api/v1/cards/{id}/comments
func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board_id query parameter is required")
		return
	}

	limit := int32(parseIntQueryParam(r, "limit", 50))
	cursor := r.URL.Query().Get("cursor")

	resp, err := h.client.ListComments(r.Context(), &commentpb.ListCommentsRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
		Limit:   limit,
		Cursor:  cursor,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"comments":    mapCommentsFromProto(resp.Comments),
		"next_cursor": resp.NextCursor,
	})
}

// UpdateComment PUT /api/v1/comments/{id}
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	commentID := r.PathValue("id")
	if commentID == "" {
		writeError(w, http.StatusBadRequest, "comment id is required")
		return
	}

	var req struct {
		BoardID string `json:"board_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateStringLen(req.Content, "content", maxContentLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	resp, err := h.client.UpdateComment(r.Context(), &commentpb.UpdateCommentRequest{
		CommentId: commentID,
		BoardId:   req.BoardID,
		UserId:    userID,
		Content:   req.Content,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"comment": mapCommentFromProto(resp.Comment),
	})
}

// DeleteComment DELETE /api/v1/comments/{id}
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	commentID := r.PathValue("id")
	if commentID == "" {
		writeError(w, http.StatusBadRequest, "comment id is required")
		return
	}

	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board_id query parameter is required")
		return
	}

	_, err := h.client.DeleteComment(r.Context(), &commentpb.DeleteCommentRequest{
		CommentId: commentID,
		BoardId:   boardID,
		UserId:    userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// GetCommentCount GET /api/v1/cards/{id}/comments/count
func (h *CommentHandler) GetCommentCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board_id query parameter is required")
		return
	}

	resp, err := h.client.GetCommentCount(r.Context(), &commentpb.GetCommentCountRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count": resp.Count,
	})
}
