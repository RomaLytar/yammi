package http

// NOTE: protoc needs to be run to regenerate Go code from board.proto before this file will compile.
// Run: cd services/api-gateway && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/v1/board/board.proto

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// ============================================================================
// Board Templates
// ============================================================================

// CreateBoardTemplate POST /api/v1/board-templates
func (h *BoardHandler) CreateBoardTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ColumnsData []struct {
			Title    string `json:"title"`
			Position int32  `json:"position"`
		} `json:"columns_data"`
		LabelsData []struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"labels_data"`
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

	var columnsData []*boardpb.BoardColumnTemplateData
	for _, col := range req.ColumnsData {
		columnsData = append(columnsData, &boardpb.BoardColumnTemplateData{
			Title:    col.Title,
			Position: col.Position,
		})
	}

	var labelsData []*boardpb.LabelTemplateData
	for _, lbl := range req.LabelsData {
		labelsData = append(labelsData, &boardpb.LabelTemplateData{
			Name:  lbl.Name,
			Color: lbl.Color,
		})
	}

	resp, err := h.client.CreateBoardTemplate(r.Context(), &boardpb.CreateBoardTemplateRequest{
		UserId:      userID,
		Name:        req.Name,
		Description: req.Description,
		ColumnsData: columnsData,
		LabelsData:  labelsData,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"template": mapBoardTemplateFromProto(resp.Template),
	})
}

// ListBoardTemplates GET /api/v1/board-templates
func (h *BoardHandler) ListBoardTemplates(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.client.ListBoardTemplates(r.Context(), &boardpb.ListBoardTemplatesRequest{
		UserId: userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"templates": mapBoardTemplatesFromProto(resp.Templates),
	})
}

// DeleteBoardTemplate DELETE /api/v1/board-templates/{id}
func (h *BoardHandler) DeleteBoardTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	templateID := r.PathValue("id")
	if templateID == "" {
		writeError(w, http.StatusBadRequest, "template id is required")
		return
	}

	_, err := h.client.DeleteBoardTemplate(r.Context(), &boardpb.DeleteBoardTemplateRequest{
		TemplateId: templateID,
		UserId:     userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// CreateBoardFromTemplate POST /api/v1/boards/from-template
func (h *BoardHandler) CreateBoardFromTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		TemplateID string `json:"template_id"`
		Title      string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.TemplateID == "" {
		writeError(w, http.StatusBadRequest, "template_id is required")
		return
	}
	if msg := validateStringLen(req.Title, "title", maxTitleLen); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	resp, err := h.client.CreateBoardFromTemplate(r.Context(), &boardpb.CreateBoardFromTemplateRequest{
		TemplateId: req.TemplateID,
		Title:      req.Title,
		UserId:     userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"board": mapBoardFromProto(resp.Board),
	})
}
