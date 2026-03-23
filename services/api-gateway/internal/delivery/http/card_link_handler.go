package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// LinkCards POST /api/v1/boards/{boardId}/cards/{cardId}/links
func (h *BoardHandler) LinkCards(w http.ResponseWriter, r *http.Request) {
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
		ChildID string `json:"child_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ChildID == "" {
		writeError(w, http.StatusBadRequest, "child_id is required")
		return
	}

	resp, err := h.client.LinkCards(r.Context(), &boardpb.LinkCardsRequest{
		BoardId:  boardID,
		UserId:   userID,
		ParentId: cardID,
		ChildId:  req.ChildID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"link": mapCardLinkFromProto(resp.Link),
	})
}

// UnlinkCards DELETE /api/v1/boards/{boardId}/card-links/{id}
func (h *BoardHandler) UnlinkCards(w http.ResponseWriter, r *http.Request) {
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

	linkID := r.PathValue("id")
	if linkID == "" {
		writeError(w, http.StatusBadRequest, "link id is required")
		return
	}

	_, err := h.client.UnlinkCards(r.Context(), &boardpb.UnlinkCardsRequest{
		LinkId:  linkID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "unlinked"})
}

// GetCardChildren GET /api/v1/boards/{boardId}/cards/{cardId}/children
func (h *BoardHandler) GetCardChildren(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetCardChildren(r.Context(), &boardpb.GetCardChildrenRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"links": mapCardLinksFromProto(resp.Links),
	})
}

// GetCardParents GET /api/v1/boards/{boardId}/cards/{cardId}/parents
func (h *BoardHandler) GetCardParents(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetCardParents(r.Context(), &boardpb.GetCardParentsRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"links": mapCardLinksFromProto(resp.Links),
	})
}
