package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// GetBoardSettings GET /api/v1/boards/{id}/settings
func (h *BoardHandler) GetBoardSettings(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetBoardSettings(r.Context(), &boardpb.GetBoardSettingsRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"settings": mapBoardSettingsFromProto(resp.Settings),
	})
}

// UpdateBoardSettings PUT /api/v1/boards/{id}/settings
func (h *BoardHandler) UpdateBoardSettings(w http.ResponseWriter, r *http.Request) {
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
		UseBoardLabelsOnly bool   `json:"use_board_labels_only"`
		DoneColumnID       string `json:"done_column_id"`
		SprintDurationDays int32  `json:"sprint_duration_days"`
		ReleasesEnabled    bool   `json:"releases_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateBoardSettings(r.Context(), &boardpb.UpdateBoardSettingsRequest{
		BoardId:            boardID,
		UserId:             userID,
		UseBoardLabelsOnly: req.UseBoardLabelsOnly,
		DoneColumnId:       req.DoneColumnID,
		SprintDurationDays: req.SprintDurationDays,
		ReleasesEnabled:    req.ReleasesEnabled,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"settings": mapBoardSettingsFromProto(resp.Settings),
	})
}

// GetAvailableLabels GET /api/v1/boards/{id}/available-labels
func (h *BoardHandler) GetAvailableLabels(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.ListAvailableLabels(r.Context(), &boardpb.ListAvailableLabelsRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"board_labels":           mapLabelsFromProto(resp.BoardLabels),
		"user_labels":            mapUserLabelsFromProto(resp.UserLabels),
		"use_board_labels_only":  resp.UseBoardLabelsOnly,
	})
}
