package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// AddMember POST /api/v1/boards/{id}/members
func (h *BoardHandler) AddMember(w http.ResponseWriter, r *http.Request) {
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
		MemberUserID string `json:"user_id"`
		Role         string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.AddMember(r.Context(), &boardpb.AddMemberRequest{
		BoardId:      boardID,
		UserId:       userID,
		MemberUserId: req.MemberUserID,
		Role:         req.Role,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"member": mapMemberFromProto(resp.Member),
	})
}

// RemoveMember DELETE /api/v1/boards/{boardId}/members/{userId}
func (h *BoardHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
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

	memberUserID := r.PathValue("userId")
	if memberUserID == "" {
		writeError(w, http.StatusBadRequest, "member user id is required")
		return
	}

	_, err := h.client.RemoveMember(r.Context(), &boardpb.RemoveMemberRequest{
		BoardId:      boardID,
		UserId:       userID,
		MemberUserId: memberUserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "removed"})
}

// ListMembers GET /api/v1/boards/{id}/members
func (h *BoardHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.ListMembers(r.Context(), &boardpb.ListMembersRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"members": mapMembersFromProto(resp.Members),
	})
}
