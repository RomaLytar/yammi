package http

import (
	"encoding/json"
	"net/http"

	authpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1"
	userpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/user"
)

type UserHandler struct {
	client     userpb.UserServiceClient
	authClient authpb.AuthServiceClient
}

func NewUserHandler(client userpb.UserServiceClient, authClient authpb.AuthServiceClient) *UserHandler {
	return &UserHandler{client: client, authClient: authClient}
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

	writeJSON(w, http.StatusOK, profileResponse{
		ID:        resp.Id,
		Email:     resp.Email,
		Name:      resp.Name,
		AvatarURL: resp.AvatarUrl,
		Bio:       resp.Bio,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user id is required")
		return
	}

	_, err := h.authClient.DeleteUser(r.Context(), &authpb.DeleteUserRequest{
		UserId: userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
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

	writeJSON(w, http.StatusOK, profileResponse{
		ID:        resp.Id,
		Email:     resp.Email,
		Name:      resp.Name,
		AvatarURL: resp.AvatarUrl,
		Bio:       resp.Bio,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	})
}
