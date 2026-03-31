package http

import (
	"encoding/json"
	"net/http"

	grpcmetadata "google.golang.org/grpc/metadata"

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
	callerID, _ := UserIDFromContext(r.Context())

	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user id is required")
		return
	}

	// Передаём caller_id в gRPC metadata для defense-in-depth проверки в auth service
	ctx := r.Context()
	if callerID != "" {
		md := grpcmetadata.Pairs("x-caller-id", callerID)
		ctx = grpcmetadata.NewOutgoingContext(ctx, md)
	}

	_, err := h.authClient.DeleteUser(ctx, &authpb.DeleteUserRequest{
		UserId: userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

func (h *UserHandler) SearchByEmail(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "q query parameter is required")
		return
	}
	if len(query) < 3 {
		writeError(w, http.StatusBadRequest, "search query must be at least 3 characters")
		return
	}

	resp, err := h.client.SearchByEmail(r.Context(), &userpb.SearchByEmailRequest{
		Query: query,
		Limit: 5,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	type userItem struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	users := make([]userItem, 0, len(resp.Users))
	for _, u := range resp.Users {
		users = append(users, userItem{
			ID:        u.Id,
			Email:     u.Email,
			Name:      u.Name,
			AvatarURL: u.AvatarUrl,
		})
	}

	writeJSON(w, http.StatusOK, struct {
		Users []userItem `json:"users"`
	}{Users: users})
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
