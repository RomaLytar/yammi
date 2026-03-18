package http

import (
	"encoding/json"
	"net/http"

	userpb "github.com/romanlovesweed/yammi/services/api-gateway/api/proto/v1/user"
)

type UserHandler struct {
	client userpb.UserServiceClient
}

func NewUserHandler(client userpb.UserServiceClient) *UserHandler {
	return &UserHandler{client: client}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user id is required")
		return
	}

	resp, err := h.client.GetProfile(r.Context(), &userpb.GetProfileRequest{
		UserId: userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"id":         resp.Id,
		"email":      resp.Email,
		"name":       resp.Name,
		"avatar_url": resp.AvatarUrl,
		"bio":        resp.Bio,
		"created_at": resp.CreatedAt,
		"updated_at": resp.UpdatedAt,
	})
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user id is required")
		return
	}

	var req struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Bio       string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateProfile(r.Context(), &userpb.UpdateProfileRequest{
		UserId:    userID,
		Name:      req.Name,
		AvatarUrl: req.AvatarURL,
		Bio:       req.Bio,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"id":         resp.Id,
		"email":      resp.Email,
		"name":       resp.Name,
		"avatar_url": resp.AvatarUrl,
		"bio":        resp.Bio,
		"created_at": resp.CreatedAt,
		"updated_at": resp.UpdatedAt,
	})
}
