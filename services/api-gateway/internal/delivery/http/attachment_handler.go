package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// CreateUploadURL POST /api/v1/cards/{id}/attachments/upload-url
func (h *BoardHandler) CreateUploadURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	var req struct {
		BoardID     string `json:"board_id"`
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
		FileSize    int64  `json:"file_size"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.BoardID == "" {
		writeError(w, http.StatusBadRequest, "board_id is required")
		return
	}
	if req.FileName == "" {
		writeError(w, http.StatusBadRequest, "file_name is required")
		return
	}
	if req.FileSize <= 0 {
		writeError(w, http.StatusBadRequest, "file_size must be positive")
		return
	}

	resp, err := h.client.CreateUploadURL(r.Context(), &boardpb.CreateUploadURLRequest{
		CardId:      cardID,
		BoardId:     req.BoardID,
		UserId:      userID,
		FileName:    req.FileName,
		ContentType: req.ContentType,
		FileSize:    req.FileSize,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"attachment": mapAttachmentFromProto(resp.Attachment),
		"upload_url": resp.UploadUrl,
	})
}

// ConfirmUpload POST /api/v1/attachments/{id}/confirm
func (h *BoardHandler) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	attachmentID := r.PathValue("id")
	if attachmentID == "" {
		writeError(w, http.StatusBadRequest, "attachment id is required")
		return
	}

	var req struct {
		BoardID string `json:"board_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.BoardID == "" {
		writeError(w, http.StatusBadRequest, "board_id is required")
		return
	}

	resp, err := h.client.ConfirmUpload(r.Context(), &boardpb.ConfirmUploadRequest{
		AttachmentId: attachmentID,
		BoardId:      req.BoardID,
		UserId:       userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"attachment": mapAttachmentFromProto(resp.Attachment),
	})
}

// GetDownloadURL GET /api/v1/attachments/{id}/download-url
func (h *BoardHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	attachmentID := r.PathValue("id")
	if attachmentID == "" {
		writeError(w, http.StatusBadRequest, "attachment id is required")
		return
	}

	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board_id query parameter is required")
		return
	}

	resp, err := h.client.GetDownloadURL(r.Context(), &boardpb.GetDownloadURLRequest{
		AttachmentId: attachmentID,
		BoardId:      boardID,
		UserId:       userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"download_url": resp.DownloadUrl,
	})
}

// ListAttachments GET /api/v1/cards/{id}/attachments
func (h *BoardHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board_id query parameter is required")
		return
	}

	resp, err := h.client.ListAttachments(r.Context(), &boardpb.ListAttachmentsRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"attachments": mapAttachmentsFromProto(resp.Attachments),
	})
}

// DeleteAttachment DELETE /api/v1/attachments/{id}
func (h *BoardHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	attachmentID := r.PathValue("id")
	if attachmentID == "" {
		writeError(w, http.StatusBadRequest, "attachment id is required")
		return
	}

	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board_id query parameter is required")
		return
	}

	_, err := h.client.DeleteAttachment(r.Context(), &boardpb.DeleteAttachmentRequest{
		AttachmentId: attachmentID,
		BoardId:      boardID,
		UserId:       userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}
