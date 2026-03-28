package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// CreateUserLabel POST /api/v1/user-labels
func (h *BoardHandler) CreateUserLabel(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateStringLen(req.Name, "name", maxNameLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if msg := validateStringLen(req.Color, "color", maxColorLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	resp, err := h.client.CreateUserLabel(r.Context(), &boardpb.CreateUserLabelRequest{
		UserId: userID,
		Name:   req.Name,
		Color:  req.Color,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"label": mapUserLabelFromProto(resp.Label),
	})
}

// ListUserLabels GET /api/v1/user-labels
func (h *BoardHandler) ListUserLabels(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.client.ListUserLabels(r.Context(), &boardpb.ListUserLabelsRequest{
		UserId: userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"labels": mapUserLabelsFromProto(resp.Labels),
	})
}

// UpdateUserLabel PUT /api/v1/user-labels/{id}
func (h *BoardHandler) UpdateUserLabel(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	labelID := r.PathValue("id")
	if labelID == "" {
		writeError(w, http.StatusBadRequest, "label id is required")
		return
	}

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateStringLen(req.Name, "name", maxNameLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if msg := validateStringLen(req.Color, "color", maxColorLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	resp, err := h.client.UpdateUserLabel(r.Context(), &boardpb.UpdateUserLabelRequest{
		LabelId: labelID,
		UserId:  userID,
		Name:    req.Name,
		Color:   req.Color,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"label": mapUserLabelFromProto(resp.Label),
	})
}

// DeleteUserLabel DELETE /api/v1/user-labels/{id}
func (h *BoardHandler) DeleteUserLabel(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	labelID := r.PathValue("id")
	if labelID == "" {
		writeError(w, http.StatusBadRequest, "label id is required")
		return
	}

	_, err := h.client.DeleteUserLabel(r.Context(), &boardpb.DeleteUserLabelRequest{
		LabelId: labelID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}
