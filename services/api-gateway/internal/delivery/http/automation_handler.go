package http

import (
	"encoding/json"
	"net/http"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

// CreateAutomationRule POST /api/v1/boards/{id}/automations
func (h *BoardHandler) CreateAutomationRule(w http.ResponseWriter, r *http.Request) {
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
		Name          string            `json:"name"`
		TriggerType   string            `json:"trigger_type"`
		TriggerConfig map[string]string `json:"trigger_config"`
		ActionType    string            `json:"action_type"`
		ActionConfig  map[string]string `json:"action_config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreateAutomationRule(r.Context(), &boardpb.CreateAutomationRuleRequest{
		BoardId:       boardID,
		UserId:        userID,
		Name:          req.Name,
		TriggerType:   req.TriggerType,
		TriggerConfig: req.TriggerConfig,
		ActionType:    req.ActionType,
		ActionConfig:  req.ActionConfig,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"rule": mapAutomationRuleFromProto(resp.Rule),
	})
}

// ListAutomationRules GET /api/v1/boards/{id}/automations
func (h *BoardHandler) ListAutomationRules(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.ListAutomationRules(r.Context(), &boardpb.ListAutomationRulesRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"rules": mapAutomationRulesFromProto(resp.Rules),
	})
}

// UpdateAutomationRule PUT /api/v1/boards/{boardId}/automations/{id}
func (h *BoardHandler) UpdateAutomationRule(w http.ResponseWriter, r *http.Request) {
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

	ruleID := r.PathValue("id")
	if ruleID == "" {
		writeError(w, http.StatusBadRequest, "rule id is required")
		return
	}

	var req struct {
		Name          string            `json:"name"`
		Enabled       bool              `json:"enabled"`
		TriggerConfig map[string]string `json:"trigger_config"`
		ActionConfig  map[string]string `json:"action_config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateAutomationRule(r.Context(), &boardpb.UpdateAutomationRuleRequest{
		RuleId:        ruleID,
		BoardId:       boardID,
		UserId:        userID,
		Name:          req.Name,
		Enabled:       req.Enabled,
		TriggerConfig: req.TriggerConfig,
		ActionConfig:  req.ActionConfig,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"rule": mapAutomationRuleFromProto(resp.Rule),
	})
}

// DeleteAutomationRule DELETE /api/v1/boards/{boardId}/automations/{id}
func (h *BoardHandler) DeleteAutomationRule(w http.ResponseWriter, r *http.Request) {
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

	ruleID := r.PathValue("id")
	if ruleID == "" {
		writeError(w, http.StatusBadRequest, "rule id is required")
		return
	}

	_, err := h.client.DeleteAutomationRule(r.Context(), &boardpb.DeleteAutomationRuleRequest{
		RuleId:  ruleID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// GetAutomationHistory GET /api/v1/boards/{boardId}/automations/{id}/history
func (h *BoardHandler) GetAutomationHistory(w http.ResponseWriter, r *http.Request) {
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

	ruleID := r.PathValue("id")
	if ruleID == "" {
		writeError(w, http.StatusBadRequest, "rule id is required")
		return
	}

	limit := parseIntQueryParam(r, "limit", 50)

	resp, err := h.client.GetAutomationHistory(r.Context(), &boardpb.GetAutomationHistoryRequest{
		RuleId:  ruleID,
		BoardId: boardID,
		UserId:  userID,
		Limit:   int32(limit),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"executions": mapAutomationExecutionsFromProto(resp.Executions),
	})
}
