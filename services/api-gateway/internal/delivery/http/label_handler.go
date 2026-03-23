package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// CreateLabel POST /api/v1/boards/{id}/labels
func (h *BoardHandler) CreateLabel(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreateLabel(r.Context(), &boardpb.CreateLabelRequest{
		BoardId: boardID,
		UserId:  userID,
		Name:    req.Name,
		Color:   req.Color,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"label": mapLabelFromProto(resp.Label),
	})
}

// ListLabels GET /api/v1/boards/{id}/labels
func (h *BoardHandler) ListLabels(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	resp, err := h.client.ListLabels(r.Context(), &boardpb.ListLabelsRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"labels": mapLabelsFromProto(resp.Labels),
	})
}

// UpdateLabel PUT /api/v1/boards/{boardId}/labels/{id}
func (h *BoardHandler) UpdateLabel(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("boardId")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	labelID := r.PathValue("id")
	if labelID == "" {
		writeError(w, http.StatusBadRequest, "label id is required")
		return
	}

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateLabel(r.Context(), &boardpb.UpdateLabelRequest{
		LabelId: labelID,
		BoardId: boardID,
		UserId:  userID,
		Name:    req.Name,
		Color:   req.Color,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"label": mapLabelFromProto(resp.Label),
	})
}

// DeleteLabel DELETE /api/v1/boards/{boardId}/labels/{id}
func (h *BoardHandler) DeleteLabel(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("boardId")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	labelID := r.PathValue("id")
	if labelID == "" {
		writeError(w, http.StatusBadRequest, "label id is required")
		return
	}

	_, err := h.client.DeleteLabel(r.Context(), &boardpb.DeleteLabelRequest{
		LabelId: labelID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// AddLabelToCard POST /api/v1/boards/{boardId}/cards/{cardId}/labels
func (h *BoardHandler) AddLabelToCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("boardId")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	cardID := r.PathValue("cardId")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	var req struct {
		LabelID string `json:"label_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.LabelID == "" {
		writeError(w, http.StatusBadRequest, "label_id is required")
		return
	}

	_, err := h.client.AddLabelToCard(r.Context(), &boardpb.AddLabelToCardRequest{
		CardId:  cardID,
		BoardId: boardID,
		LabelId: req.LabelID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "added"})
}

// RemoveLabelFromCard DELETE /api/v1/boards/{boardId}/cards/{cardId}/labels/{labelId}
func (h *BoardHandler) RemoveLabelFromCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("boardId")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	cardID := r.PathValue("cardId")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	labelID := r.PathValue("labelId")
	if labelID == "" {
		writeError(w, http.StatusBadRequest, "label id is required")
		return
	}

	_, err := h.client.RemoveLabelFromCard(r.Context(), &boardpb.RemoveLabelFromCardRequest{
		CardId:  cardID,
		BoardId: boardID,
		LabelId: labelID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "removed"})
}

// GetCardLabels GET /api/v1/boards/{boardId}/cards/{cardId}/labels
func (h *BoardHandler) GetCardLabels(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("boardId")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	cardID := r.PathValue("cardId")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	resp, err := h.client.GetCardLabels(r.Context(), &boardpb.GetCardLabelsRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"labels": mapLabelsFromProto(resp.Labels),
	})
}
