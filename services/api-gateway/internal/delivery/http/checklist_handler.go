package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// CreateChecklist POST /api/v1/boards/{boardId}/cards/{cardId}/checklists
func (h *BoardHandler) CreateChecklist(w http.ResponseWriter, r *http.Request) {
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
		Title    string `json:"title"`
		Position int32  `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreateChecklist(r.Context(), &boardpb.CreateChecklistRequest{
		CardId:   cardID,
		BoardId:  boardID,
		UserId:   userID,
		Title:    req.Title,
		Position: req.Position,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"checklist": mapChecklistFromProto(resp.Checklist),
	})
}

// GetChecklists GET /api/v1/boards/{boardId}/cards/{cardId}/checklists
func (h *BoardHandler) GetChecklists(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetChecklists(r.Context(), &boardpb.GetChecklistsRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"checklists": mapChecklistsFromProto(resp.Checklists),
	})
}

// UpdateChecklist PUT /api/v1/boards/{boardId}/checklists/{id}
func (h *BoardHandler) UpdateChecklist(w http.ResponseWriter, r *http.Request) {
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

	checklistID := r.PathValue("id")
	if checklistID == "" {
		writeError(w, http.StatusBadRequest, "checklist id is required")
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateChecklist(r.Context(), &boardpb.UpdateChecklistRequest{
		ChecklistId: checklistID,
		BoardId:     boardID,
		UserId:      userID,
		Title:       req.Title,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"checklist": mapChecklistFromProto(resp.Checklist),
	})
}

// DeleteChecklist DELETE /api/v1/boards/{boardId}/checklists/{id}
func (h *BoardHandler) DeleteChecklist(w http.ResponseWriter, r *http.Request) {
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

	checklistID := r.PathValue("id")
	if checklistID == "" {
		writeError(w, http.StatusBadRequest, "checklist id is required")
		return
	}

	_, err := h.client.DeleteChecklist(r.Context(), &boardpb.DeleteChecklistRequest{
		ChecklistId: checklistID,
		BoardId:     boardID,
		UserId:      userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// CreateChecklistItem POST /api/v1/boards/{boardId}/checklists/{checklistId}/items
func (h *BoardHandler) CreateChecklistItem(w http.ResponseWriter, r *http.Request) {
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

	checklistID := r.PathValue("checklistId")
	if checklistID == "" {
		writeError(w, http.StatusBadRequest, "checklist id is required")
		return
	}

	var req struct {
		Title    string `json:"title"`
		Position int32  `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreateChecklistItem(r.Context(), &boardpb.CreateChecklistItemRequest{
		ChecklistId: checklistID,
		BoardId:     boardID,
		UserId:      userID,
		Title:       req.Title,
		Position:    req.Position,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"item": mapChecklistItemFromProto(resp.Item),
	})
}

// UpdateChecklistItem PUT /api/v1/boards/{boardId}/checklist-items/{id}
func (h *BoardHandler) UpdateChecklistItem(w http.ResponseWriter, r *http.Request) {
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

	itemID := r.PathValue("id")
	if itemID == "" {
		writeError(w, http.StatusBadRequest, "item id is required")
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateChecklistItem(r.Context(), &boardpb.UpdateChecklistItemRequest{
		ItemId:  itemID,
		BoardId: boardID,
		UserId:  userID,
		Title:   req.Title,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"item": mapChecklistItemFromProto(resp.Item),
	})
}

// DeleteChecklistItem DELETE /api/v1/boards/{boardId}/checklist-items/{id}
func (h *BoardHandler) DeleteChecklistItem(w http.ResponseWriter, r *http.Request) {
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

	itemID := r.PathValue("id")
	if itemID == "" {
		writeError(w, http.StatusBadRequest, "item id is required")
		return
	}

	_, err := h.client.DeleteChecklistItem(r.Context(), &boardpb.DeleteChecklistItemRequest{
		ItemId:  itemID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// ToggleChecklistItem PUT /api/v1/boards/{boardId}/checklist-items/{id}/toggle
func (h *BoardHandler) ToggleChecklistItem(w http.ResponseWriter, r *http.Request) {
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

	itemID := r.PathValue("id")
	if itemID == "" {
		writeError(w, http.StatusBadRequest, "item id is required")
		return
	}

	resp, err := h.client.ToggleChecklistItem(r.Context(), &boardpb.ToggleChecklistItemRequest{
		ItemId:  itemID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"is_checked": resp.IsChecked,
	})
}
