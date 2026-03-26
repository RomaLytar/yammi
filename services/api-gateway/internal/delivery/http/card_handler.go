package http

import (
	"encoding/json"
	"net/http"
	"time"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateCard POST /api/v1/columns/{id}/cards
func (h *BoardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	columnID := r.PathValue("id")
	if columnID == "" {
		writeError(w, http.StatusBadRequest, "column id is required")
		return
	}

	var req struct {
		BoardID     string  `json:"board_id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Position    string  `json:"position"`
		AssigneeID  string  `json:"assignee_id"`
		DueDate     *string `json:"due_date"`
		Priority    string  `json:"priority"`
		TaskType    string  `json:"task_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateStringLen(req.Title, "title", maxTitleLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if msg := validateStringLen(req.Description, "description", maxDescriptionLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	grpcReq := &boardpb.CreateCardRequest{
		ColumnId:    columnID,
		BoardId:     req.BoardID,
		UserId:      userID,
		Title:       req.Title,
		Description: req.Description,
		Position:    req.Position,
		AssigneeId:  req.AssigneeID,
		Priority:    req.Priority,
		TaskType:    req.TaskType,
	}
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid due_date format, expected RFC3339")
			return
		}
		grpcReq.DueDate = timestamppb.New(t)
	}

	resp, err := h.client.CreateCard(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"card": mapCardFromProto(resp.Card),
	})
}

// GetCard GET /api/v1/cards/{id}
func (h *BoardHandler) GetCard(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetCard(r.Context(), &boardpb.GetCardRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"card": mapCardFromProto(resp.Card),
	})
}

// GetCards GET /api/v1/columns/{id}/cards
func (h *BoardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	columnID := r.PathValue("id")
	if columnID == "" {
		writeError(w, http.StatusBadRequest, "column id is required")
		return
	}

	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board_id query parameter is required")
		return
	}

	resp, err := h.client.GetCards(r.Context(), &boardpb.GetCardsRequest{
		ColumnId: columnID,
		BoardId:  boardID,
		UserId:   userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cards": mapCardsFromProto(resp.Cards),
	})
}

// UpdateCard PUT /api/v1/cards/{id}
func (h *BoardHandler) UpdateCard(w http.ResponseWriter, r *http.Request) {
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
		BoardID     string  `json:"board_id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		AssigneeID  string  `json:"assignee_id"`
		Version     int32   `json:"version"`
		DueDate     *string `json:"due_date"`
		Priority    string  `json:"priority"`
		TaskType    string  `json:"task_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateStringLen(req.Title, "title", maxTitleLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if msg := validateStringLen(req.Description, "description", maxDescriptionLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	grpcReq := &boardpb.UpdateCardRequest{
		CardId:      cardID,
		BoardId:     req.BoardID,
		UserId:      userID,
		Title:       req.Title,
		Description: req.Description,
		AssigneeId:  req.AssigneeID,
		Version:     req.Version,
		Priority:    req.Priority,
		TaskType:    req.TaskType,
	}
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid due_date format, expected RFC3339")
			return
		}
		grpcReq.DueDate = timestamppb.New(t)
	}

	resp, err := h.client.UpdateCard(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"card": mapCardFromProto(resp.Card),
	})
}

// MoveCard PUT /api/v1/cards/{id}/move
func (h *BoardHandler) MoveCard(w http.ResponseWriter, r *http.Request) {
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
		BoardID      string `json:"board_id"`
		FromColumnID string `json:"from_column_id"`
		ToColumnID   string `json:"to_column_id"`
		Position     string `json:"position"`
		Version      int32  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.MoveCard(r.Context(), &boardpb.MoveCardRequest{
		CardId:       cardID,
		BoardId:      req.BoardID,
		FromColumnId: req.FromColumnID,
		ToColumnId:   req.ToColumnID,
		Position:     req.Position,
		UserId:       userID,
		Version:      req.Version,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"card":            mapCardFromProto(resp.Card),
		"cards_in_column": mapCardsFromProto(resp.CardsInColumn),
	})
}

// DeleteCards POST /api/v1/cards/delete — удаление одной или нескольких карточек
func (h *BoardHandler) DeleteCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		CardIDs []string `json:"card_ids"`
		BoardID string   `json:"board_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.CardIDs) == 0 {
		writeError(w, http.StatusBadRequest, "card_ids is required")
		return
	}
	if len(req.CardIDs) > 100 {
		writeError(w, http.StatusBadRequest, "too many card_ids, max 100")
		return
	}
	if req.BoardID == "" {
		writeError(w, http.StatusBadRequest, "board_id is required")
		return
	}

	_, err := h.client.DeleteCard(r.Context(), &boardpb.DeleteCardRequest{
		CardIds: req.CardIDs,
		BoardId: req.BoardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// AssignCard PUT /api/v1/cards/{id}/assign
func (h *BoardHandler) AssignCard(w http.ResponseWriter, r *http.Request) {
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
		BoardID    string `json:"board_id"`
		AssigneeID string `json:"assignee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.BoardID == "" {
		writeError(w, http.StatusBadRequest, "board_id is required")
		return
	}
	if req.AssigneeID == "" {
		writeError(w, http.StatusBadRequest, "assignee_id is required")
		return
	}

	resp, err := h.client.AssignCard(r.Context(), &boardpb.AssignCardRequest{
		CardId:     cardID,
		BoardId:    req.BoardID,
		UserId:     userID,
		AssigneeId: req.AssigneeID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"card": mapCardFromProto(resp.Card),
	})
}

// GetCardActivity GET /api/v1/cards/{id}/activity
func (h *BoardHandler) GetCardActivity(w http.ResponseWriter, r *http.Request) {
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

	limit := int32(parseIntQueryParam(r, "limit", 20))
	cursor := r.URL.Query().Get("cursor")

	resp, err := h.client.GetCardActivity(r.Context(), &boardpb.GetCardActivityRequest{
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

	entries := make([]map[string]interface{}, len(resp.Entries))
	for i, e := range resp.Entries {
		entry := map[string]interface{}{
			"id":            e.Id,
			"card_id":       e.CardId,
			"board_id":      e.BoardId,
			"actor_id":      e.ActorId,
			"activity_type": e.ActivityType,
			"description":   e.Description,
			"changes":       e.Changes,
		}
		if e.CreatedAt != nil {
			entry["created_at"] = e.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
		}
		entries[i] = entry
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"entries":     entries,
		"next_cursor": resp.NextCursor,
	})
}

// UnassignCard DELETE /api/v1/cards/{id}/assign
func (h *BoardHandler) UnassignCard(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.UnassignCard(r.Context(), &boardpb.UnassignCardRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"card": mapCardFromProto(resp.Card),
	})
}
