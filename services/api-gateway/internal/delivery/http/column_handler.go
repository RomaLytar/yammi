package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// AddColumn POST /api/v1/boards/{id}/columns
func (h *BoardHandler) AddColumn(w http.ResponseWriter, r *http.Request) {
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
		Title    string `json:"title"`
		Position int32  `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.AddColumn(r.Context(), &boardpb.AddColumnRequest{
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
		"column": mapColumnFromProto(resp.Column),
	})
}

// GetColumns GET /api/v1/boards/{id}/columns
func (h *BoardHandler) GetColumns(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetColumns(r.Context(), &boardpb.GetColumnsRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"columns": mapColumnsFromProto(resp.Columns),
	})
}

// UpdateColumn PUT /api/v1/columns/{id}
func (h *BoardHandler) UpdateColumn(w http.ResponseWriter, r *http.Request) {
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
		BoardID string `json:"board_id"`
		Title   string `json:"title"`
		Version int32  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateColumn(r.Context(), &boardpb.UpdateColumnRequest{
		ColumnId: columnID,
		BoardId:  req.BoardID,
		UserId:   userID,
		Title:    req.Title,
		Version:  req.Version,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"column": mapColumnFromProto(resp.Column),
	})
}

// DeleteColumn DELETE /api/v1/columns/{id}
func (h *BoardHandler) DeleteColumn(w http.ResponseWriter, r *http.Request) {
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
		BoardID string `json:"board_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := h.client.DeleteColumn(r.Context(), &boardpb.DeleteColumnRequest{
		ColumnId: columnID,
		BoardId:  req.BoardID,
		UserId:   userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// ReorderColumns PUT /api/v1/boards/{id}/columns/reorder
func (h *BoardHandler) ReorderColumns(w http.ResponseWriter, r *http.Request) {
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
		Positions []struct {
			ColumnID string `json:"column_id"`
			Position int32  `json:"position"`
		} `json:"positions"`
		Version int32 `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	positions := make([]*boardpb.ColumnPosition, len(req.Positions))
	for i, p := range req.Positions {
		positions[i] = &boardpb.ColumnPosition{
			ColumnId: p.ColumnID,
			Position: p.Position,
		}
	}

	resp, err := h.client.ReorderColumns(r.Context(), &boardpb.ReorderColumnsRequest{
		BoardId:   boardID,
		UserId:    userID,
		Positions: positions,
		Version:   req.Version,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"columns": mapColumnsFromProto(resp.Columns),
	})
}
