package http

import (
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// CreateCustomField POST /api/v1/boards/{id}/custom-fields
func (h *BoardHandler) CreateCustomField(w http.ResponseWriter, r *http.Request) {
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
		Name      string   `json:"name"`
		FieldType string   `json:"field_type"`
		Options   []string `json:"options"`
		Position  int32    `json:"position"`
		Required  bool     `json:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreateCustomField(r.Context(), &boardpb.CreateCustomFieldRequest{
		BoardId:   boardID,
		UserId:    userID,
		Name:      req.Name,
		FieldType: req.FieldType,
		Options:   req.Options,
		Position:  req.Position,
		Required:  req.Required,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"definition": mapCustomFieldDefFromProto(resp.Definition),
	})
}

// ListCustomFields GET /api/v1/boards/{id}/custom-fields
func (h *BoardHandler) ListCustomFields(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.ListCustomFields(r.Context(), &boardpb.ListCustomFieldsRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"definitions": mapCustomFieldDefsFromProto(resp.Definitions),
	})
}

// UpdateCustomField PUT /api/v1/boards/{boardId}/custom-fields/{id}
func (h *BoardHandler) UpdateCustomField(w http.ResponseWriter, r *http.Request) {
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

	fieldID := r.PathValue("id")
	if fieldID == "" {
		writeError(w, http.StatusBadRequest, "field id is required")
		return
	}

	var req struct {
		Name     string   `json:"name"`
		Options  []string `json:"options"`
		Required bool     `json:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateCustomField(r.Context(), &boardpb.UpdateCustomFieldRequest{
		FieldId:  fieldID,
		BoardId:  boardID,
		UserId:   userID,
		Name:     req.Name,
		Options:  req.Options,
		Required: req.Required,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"definition": mapCustomFieldDefFromProto(resp.Definition),
	})
}

// DeleteCustomField DELETE /api/v1/boards/{boardId}/custom-fields/{id}
func (h *BoardHandler) DeleteCustomField(w http.ResponseWriter, r *http.Request) {
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

	fieldID := r.PathValue("id")
	if fieldID == "" {
		writeError(w, http.StatusBadRequest, "field id is required")
		return
	}

	_, err := h.client.DeleteCustomField(r.Context(), &boardpb.DeleteCustomFieldRequest{
		FieldId: fieldID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// SetCustomFieldValue PUT /api/v1/boards/{boardId}/cards/{cardId}/custom-fields/{fieldId}
func (h *BoardHandler) SetCustomFieldValue(w http.ResponseWriter, r *http.Request) {
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

	fieldID := r.PathValue("fieldId")
	if fieldID == "" {
		writeError(w, http.StatusBadRequest, "field id is required")
		return
	}

	var req struct {
		ValueText   *string  `json:"value_text"`
		ValueNumber *float64 `json:"value_number"`
		ValueDate   *string  `json:"value_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pbReq := &boardpb.SetCustomFieldValueRequest{
		CardId:  cardID,
		BoardId: boardID,
		FieldId: fieldID,
		UserId:  userID,
	}

	if req.ValueText != nil {
		pbReq.ValueText = *req.ValueText
		pbReq.HasText = true
	}
	if req.ValueNumber != nil {
		pbReq.ValueNumber = *req.ValueNumber
		pbReq.HasNumber = true
	}
	if req.ValueDate != nil {
		t, err := parseTime(*req.ValueDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, use RFC3339")
			return
		}
		pbReq.ValueDate = timestamppb.New(t)
		pbReq.HasDate = true
	}

	resp, err := h.client.SetCustomFieldValue(r.Context(), pbReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"value": mapCustomFieldValueFromProto(resp.Value),
	})
}

// GetCardCustomFields GET /api/v1/boards/{boardId}/cards/{cardId}/custom-fields
func (h *BoardHandler) GetCardCustomFields(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.GetCardCustomFields(r.Context(), &boardpb.GetCardCustomFieldsRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"values": mapCustomFieldValuesFromProto(resp.Values),
	})
}
