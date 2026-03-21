package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

type BoardHandler struct {
	client boardpb.BoardServiceClient
}

func NewBoardHandler(client boardpb.BoardServiceClient) *BoardHandler {
	return &BoardHandler{client: client}
}

// CreateBoard POST /api/v1/boards
func (h *BoardHandler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreateBoard(r.Context(), &boardpb.CreateBoardRequest{
		Title:       req.Title,
		Description: req.Description,
		OwnerId:     userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"board": mapBoardFromProto(resp.Board),
	})
}

// GetBoard GET /api/v1/boards/{id}
func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetBoard(r.Context(), &boardpb.GetBoardRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"board":   mapBoardFromProto(resp.Board),
		"columns": mapColumnsFromProto(resp.Columns),
		"members": mapMembersFromProto(resp.Members),
	})
}

// ListBoards GET /api/v1/boards
func (h *BoardHandler) ListBoards(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseIntQueryParam(r, "limit", 20)
	cursor := r.URL.Query().Get("cursor")
	ownerOnly := r.URL.Query().Get("owner_only") == "true"
	search := r.URL.Query().Get("search")
	sortBy := r.URL.Query().Get("sort_by")

	resp, err := h.client.ListBoards(r.Context(), &boardpb.ListBoardsRequest{
		UserId:    userID,
		Limit:     int32(limit),
		Cursor:    cursor,
		OwnerOnly: ownerOnly,
		Search:    search,
		SortBy:    sortBy,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"boards":      mapBoardsFromProto(resp.Boards),
		"next_cursor": resp.NextCursor,
	})
}

// UpdateBoard PUT /api/v1/boards/{id}
func (h *BoardHandler) UpdateBoard(w http.ResponseWriter, r *http.Request) {
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
		Title       string `json:"title"`
		Description string `json:"description"`
		Version     int32  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateBoard(r.Context(), &boardpb.UpdateBoardRequest{
		BoardId:     boardID,
		UserId:      userID,
		Title:       req.Title,
		Description: req.Description,
		Version:     req.Version,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"board": mapBoardFromProto(resp.Board),
	})
}

// DeleteBoards POST /api/v1/boards/delete — удаление одной или нескольких досок
func (h *BoardHandler) DeleteBoards(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		BoardIDs []string `json:"board_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.BoardIDs) == 0 {
		writeError(w, http.StatusBadRequest, "board_ids is required")
		return
	}

	_, err := h.client.DeleteBoard(r.Context(), &boardpb.DeleteBoardRequest{
		BoardIds: req.BoardIDs,
		UserId:   userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}
