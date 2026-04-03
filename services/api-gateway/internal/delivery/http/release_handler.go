package http

import (
	"encoding/json"
	"net/http"
	"time"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateRelease POST /api/v1/boards/{id}/releases
func (h *BoardHandler) CreateRelease(w http.ResponseWriter, r *http.Request) {
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
		Name        string  `json:"name"`
		Description string  `json:"description"`
		StartDate   *string `json:"start_date"`
		EndDate     *string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateStringLen(req.Name, "name", maxNameLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if msg := validateStringLen(req.Description, "description", maxDescriptionLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	grpcReq := &boardpb.CreateReleaseRequest{
		BoardId:     boardID,
		UserId:      userID,
		Name:        req.Name,
		Description: req.Description,
	}
	if req.StartDate != nil && *req.StartDate != "" {
		t, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid start_date format, expected RFC3339")
			return
		}
		grpcReq.StartDate = timestamppb.New(t)
	}
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_date format, expected RFC3339")
			return
		}
		grpcReq.EndDate = timestamppb.New(t)
	}

	resp, err := h.client.CreateRelease(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"release": mapReleaseFromProto(resp.Release),
	})
}

// GetRelease GET /api/v1/boards/{boardId}/releases/{releaseId}
func (h *BoardHandler) GetRelease(w http.ResponseWriter, r *http.Request) {
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
	releaseID := r.PathValue("releaseId")
	if releaseID == "" {
		writeError(w, http.StatusBadRequest, "release id is required")
		return
	}

	resp, err := h.client.GetRelease(r.Context(), &boardpb.GetReleaseRequest{
		ReleaseId: releaseID,
		BoardId:   boardID,
		UserId:    userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"release": mapReleaseFromProto(resp.Release),
	})
}

// ListReleases GET /api/v1/boards/{id}/releases
func (h *BoardHandler) ListReleases(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.ListReleases(r.Context(), &boardpb.ListReleasesRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"releases": mapReleasesFromProto(resp.Releases),
	})
}

// UpdateRelease PUT /api/v1/boards/{boardId}/releases/{releaseId}
func (h *BoardHandler) UpdateRelease(w http.ResponseWriter, r *http.Request) {
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
	releaseID := r.PathValue("releaseId")
	if releaseID == "" {
		writeError(w, http.StatusBadRequest, "release id is required")
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Version     int32   `json:"version"`
		StartDate   *string `json:"start_date"`
		EndDate     *string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateStringLen(req.Name, "name", maxNameLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if msg := validateStringLen(req.Description, "description", maxDescriptionLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	grpcReq := &boardpb.UpdateReleaseRequest{
		ReleaseId:   releaseID,
		BoardId:     boardID,
		UserId:      userID,
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
	}
	if req.StartDate != nil && *req.StartDate != "" {
		t, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid start_date format, expected RFC3339")
			return
		}
		grpcReq.StartDate = timestamppb.New(t)
	}
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_date format, expected RFC3339")
			return
		}
		grpcReq.EndDate = timestamppb.New(t)
	}

	resp, err := h.client.UpdateRelease(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"release": mapReleaseFromProto(resp.Release),
	})
}

// DeleteRelease DELETE /api/v1/boards/{boardId}/releases/{releaseId}
func (h *BoardHandler) DeleteRelease(w http.ResponseWriter, r *http.Request) {
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
	releaseID := r.PathValue("releaseId")
	if releaseID == "" {
		writeError(w, http.StatusBadRequest, "release id is required")
		return
	}

	_, err := h.client.DeleteRelease(r.Context(), &boardpb.DeleteReleaseRequest{
		ReleaseId: releaseID,
		BoardId:   boardID,
		UserId:    userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// StartRelease POST /api/v1/boards/{boardId}/releases/{releaseId}/start
func (h *BoardHandler) StartRelease(w http.ResponseWriter, r *http.Request) {
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
	releaseID := r.PathValue("releaseId")
	if releaseID == "" {
		writeError(w, http.StatusBadRequest, "release id is required")
		return
	}

	resp, err := h.client.StartRelease(r.Context(), &boardpb.StartReleaseRequest{
		ReleaseId: releaseID,
		BoardId:   boardID,
		UserId:    userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"release": mapReleaseFromProto(resp.Release),
	})
}

// CompleteRelease POST /api/v1/boards/{boardId}/releases/{releaseId}/complete
func (h *BoardHandler) CompleteRelease(w http.ResponseWriter, r *http.Request) {
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
	releaseID := r.PathValue("releaseId")
	if releaseID == "" {
		writeError(w, http.StatusBadRequest, "release id is required")
		return
	}

	resp, err := h.client.CompleteRelease(r.Context(), &boardpb.CompleteReleaseRequest{
		ReleaseId: releaseID,
		BoardId:   boardID,
		UserId:    userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"release":                mapReleaseFromProto(resp.Release),
		"cards_moved_to_backlog": resp.CardsMovedToBacklog,
	})
}

// GetActiveRelease GET /api/v1/boards/{id}/releases/active
func (h *BoardHandler) GetActiveRelease(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetActiveRelease(r.Context(), &boardpb.GetActiveReleaseRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	if resp.Release != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"release": mapReleaseFromProto(resp.Release),
		})
	} else {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"release": nil,
		})
	}
}

// AssignCardToRelease POST /api/v1/boards/{boardId}/releases/{releaseId}/cards
func (h *BoardHandler) AssignCardToRelease(w http.ResponseWriter, r *http.Request) {
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
	releaseID := r.PathValue("releaseId")
	if releaseID == "" {
		writeError(w, http.StatusBadRequest, "release id is required")
		return
	}

	var req struct {
		CardID string `json:"card_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CardID == "" {
		writeError(w, http.StatusBadRequest, "card_id is required")
		return
	}

	_, err := h.client.AssignCardToRelease(r.Context(), &boardpb.AssignCardToReleaseRequest{
		BoardId:   boardID,
		UserId:    userID,
		ReleaseId: releaseID,
		CardId:    req.CardID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "assigned"})
}

// RemoveCardFromRelease DELETE /api/v1/boards/{boardId}/releases/{releaseId}/cards/{cardId}
func (h *BoardHandler) RemoveCardFromRelease(w http.ResponseWriter, r *http.Request) {
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
	releaseID := r.PathValue("releaseId")
	if releaseID == "" {
		writeError(w, http.StatusBadRequest, "release id is required")
		return
	}
	cardID := r.PathValue("cardId")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	_, err := h.client.RemoveCardFromRelease(r.Context(), &boardpb.RemoveCardFromReleaseRequest{
		BoardId:   boardID,
		UserId:    userID,
		ReleaseId: releaseID,
		CardId:    cardID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "removed"})
}

// GetBacklog GET /api/v1/boards/{id}/backlog
func (h *BoardHandler) GetBacklog(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetBacklog(r.Context(), &boardpb.GetBacklogRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cards": mapCardsFromProto(resp.Cards),
	})
}

// GetReleaseCards GET /api/v1/boards/{boardId}/releases/{releaseId}/cards
func (h *BoardHandler) GetReleaseCards(w http.ResponseWriter, r *http.Request) {
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
	releaseID := r.PathValue("releaseId")
	if releaseID == "" {
		writeError(w, http.StatusBadRequest, "release id is required")
		return
	}

	resp, err := h.client.GetReleaseCards(r.Context(), &boardpb.GetReleaseCardsRequest{
		BoardId:   boardID,
		UserId:    userID,
		ReleaseId: releaseID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cards": mapCardsFromProto(resp.Cards),
	})
}
