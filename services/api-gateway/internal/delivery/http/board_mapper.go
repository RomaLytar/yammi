package http

import (
	"net/http"
	"strconv"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
	userpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/user"
)

func mapBoardFromProto(pb *boardpb.Board) boardResponse {
	return boardResponse{
		ID:          pb.Id,
		Title:       pb.Title,
		Description: pb.Description,
		OwnerID:     pb.OwnerId,
		Version:     pb.Version,
		CreatedAt:   pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapBoardsFromProto(pbs []*boardpb.Board) []boardResponse {
	boards := make([]boardResponse, len(pbs))
	for i, pb := range pbs {
		boards[i] = mapBoardFromProto(pb)
	}
	return boards
}

func mapBoardsWithOwners(pbs []*boardpb.Board, owners map[string]*userpb.UserInfo) []boardResponse {
	boards := make([]boardResponse, len(pbs))
	for i, pb := range pbs {
		b := mapBoardFromProto(pb)
		if owner, ok := owners[pb.OwnerId]; ok {
			b.OwnerName = owner.Name
			b.OwnerAvatarURL = owner.AvatarUrl
		}
		boards[i] = b
	}
	return boards
}

func mapColumnFromProto(pb *boardpb.Column) columnResponse {
	return columnResponse{
		ID:        pb.Id,
		BoardID:   pb.BoardId,
		Title:     pb.Title,
		Position:  pb.Position,
		Version:   pb.Version,
		CreatedAt: pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		CardCount: pb.CardCount,
	}
}

func mapColumnsFromProto(pbs []*boardpb.Column) []columnResponse {
	columns := make([]columnResponse, len(pbs))
	for i, pb := range pbs {
		columns[i] = mapColumnFromProto(pb)
	}
	return columns
}

func mapCardFromProto(pb *boardpb.Card) cardResponse {
	resp := cardResponse{
		ID:          pb.Id,
		ColumnID:    pb.ColumnId,
		BoardID:     pb.BoardId,
		Title:       pb.Title,
		Description: pb.Description,
		Position:    pb.Position,
		AssigneeID:  pb.AssigneeId,
		CreatorID:   pb.CreatorId,
		Version:     pb.Version,
		CreatedAt:   pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		Priority:    pb.Priority,
		TaskType:    pb.TaskType,
	}
	if pb.DueDate != nil {
		resp.DueDate = pb.DueDate.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	return resp
}

func mapCardsFromProto(pbs []*boardpb.Card) []cardResponse {
	cards := make([]cardResponse, len(pbs))
	for i, pb := range pbs {
		cards[i] = mapCardFromProto(pb)
	}
	return cards
}

func mapMemberFromProto(pb *boardpb.BoardMember) memberResponse {
	return memberResponse{
		UserID:   pb.UserId,
		Role:     pb.Role,
		Version:  pb.Version,
		JoinedAt: pb.JoinedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapMembersFromProto(pbs []*boardpb.BoardMember) []memberResponse {
	members := make([]memberResponse, len(pbs))
	for i, pb := range pbs {
		members[i] = mapMemberFromProto(pb)
	}
	return members
}

func mapMembersWithProfiles(pbs []*boardpb.BoardMember, profiles map[string]*userpb.UserInfo) []memberResponse {
	members := make([]memberResponse, len(pbs))
	for i, pb := range pbs {
		m := mapMemberFromProto(pb)
		if profile, ok := profiles[pb.UserId]; ok {
			m.Name = profile.Name
			m.Email = profile.Email
			m.AvatarURL = profile.AvatarUrl
		}
		members[i] = m
	}
	return members
}

func mapAttachmentFromProto(pb *boardpb.Attachment) attachmentResponse {
	return attachmentResponse{
		ID:         pb.Id,
		CardID:     pb.CardId,
		BoardID:    pb.BoardId,
		FileName:   pb.FileName,
		FileSize:   pb.FileSize,
		MimeType:   pb.MimeType,
		UploaderID: pb.UploaderId,
		CreatedAt:  pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapAttachmentsFromProto(pbs []*boardpb.Attachment) []attachmentResponse {
	attachments := make([]attachmentResponse, len(pbs))
	for i, pb := range pbs {
		attachments[i] = mapAttachmentFromProto(pb)
	}
	return attachments
}

func mapLabelFromProto(pb *boardpb.Label) labelResponse {
	return labelResponse{
		ID:        pb.Id,
		BoardID:   pb.BoardId,
		Name:      pb.Name,
		Color:     pb.Color,
		CreatedAt: pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapLabelsFromProto(pbs []*boardpb.Label) []labelResponse {
	labels := make([]labelResponse, len(pbs))
	for i, pb := range pbs {
		labels[i] = mapLabelFromProto(pb)
	}
	return labels
}

func mapCardLinkFromProto(pb *boardpb.CardLink) cardLinkResponse {
	return cardLinkResponse{
		ID:              pb.Id,
		ParentID:        pb.ParentId,
		ChildID:         pb.ChildId,
		BoardID:         pb.BoardId,
		LinkType:        pb.LinkType,
		ChildTitle:      pb.ChildTitle,
		ChildColumnName: pb.ChildColumnName,
		CreatedAt:       pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapCardLinksFromProto(pbs []*boardpb.CardLink) []cardLinkResponse {
	links := make([]cardLinkResponse, len(pbs))
	for i, pb := range pbs {
		links[i] = mapCardLinkFromProto(pb)
	}
	return links
}

func mapCustomFieldDefFromProto(pb *boardpb.CustomFieldDefinition) customFieldDefResponse {
	return customFieldDefResponse{
		ID:        pb.Id,
		BoardID:   pb.BoardId,
		Name:      pb.Name,
		FieldType: pb.FieldType,
		Options:   pb.Options,
		Position:  pb.Position,
		Required:  pb.Required,
		CreatedAt: pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapCustomFieldDefsFromProto(pbs []*boardpb.CustomFieldDefinition) []customFieldDefResponse {
	defs := make([]customFieldDefResponse, len(pbs))
	for i, pb := range pbs {
		defs[i] = mapCustomFieldDefFromProto(pb)
	}
	return defs
}

func mapCustomFieldValueFromProto(pb *boardpb.CustomFieldValue) customFieldValueResponse {
	resp := customFieldValueResponse{
		ID:        pb.Id,
		CardID:    pb.CardId,
		BoardID:   pb.BoardId,
		FieldID:   pb.FieldId,
		CreatedAt: pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
	if pb.HasText {
		resp.ValueText = &pb.ValueText
	}
	if pb.HasNumber {
		resp.ValueNumber = &pb.ValueNumber
	}
	if pb.HasDate && pb.ValueDate != nil {
		dateStr := pb.ValueDate.AsTime().Format("2006-01-02T15:04:05Z07:00")
		resp.ValueDate = &dateStr
	}
	return resp
}

func mapCustomFieldValuesFromProto(pbs []*boardpb.CustomFieldValue) []customFieldValueResponse {
	values := make([]customFieldValueResponse, len(pbs))
	for i, pb := range pbs {
		values[i] = mapCustomFieldValueFromProto(pb)
	}
	return values
}

func mapAutomationRuleFromProto(pb *boardpb.AutomationRule) automationRuleResponse {
	return automationRuleResponse{
		ID:            pb.Id,
		BoardID:       pb.BoardId,
		Name:          pb.Name,
		Enabled:       pb.Enabled,
		TriggerType:   pb.TriggerType,
		TriggerConfig: pb.TriggerConfig,
		ActionType:    pb.ActionType,
		ActionConfig:  pb.ActionConfig,
		CreatedBy:     pb.CreatedBy,
		CreatedAt:     pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapAutomationRulesFromProto(pbs []*boardpb.AutomationRule) []automationRuleResponse {
	rules := make([]automationRuleResponse, len(pbs))
	for i, pb := range pbs {
		rules[i] = mapAutomationRuleFromProto(pb)
	}
	return rules
}

func mapAutomationExecutionFromProto(pb *boardpb.AutomationExecution) automationExecutionResponse {
	return automationExecutionResponse{
		ID:             pb.Id,
		RuleID:         pb.RuleId,
		BoardID:        pb.BoardId,
		CardID:         pb.CardId,
		TriggerEventID: pb.TriggerEventId,
		Status:         pb.Status,
		ErrorMessage:   pb.ErrorMessage,
		ExecutedAt:     pb.ExecutedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapAutomationExecutionsFromProto(pbs []*boardpb.AutomationExecution) []automationExecutionResponse {
	execs := make([]automationExecutionResponse, len(pbs))
	for i, pb := range pbs {
		execs[i] = mapAutomationExecutionFromProto(pb)
	}
	return execs
}

func mapChecklistFromProto(pb *boardpb.Checklist) checklistResponse {
	items := make([]checklistItemResponse, 0, len(pb.GetItems()))
	for _, item := range pb.GetItems() {
		items = append(items, mapChecklistItemFromProto(item))
	}
	return checklistResponse{
		ID:        pb.GetId(),
		CardID:    pb.GetCardId(),
		BoardID:   pb.GetBoardId(),
		Title:     pb.GetTitle(),
		Position:  pb.GetPosition(),
		Items:     items,
		Progress:  pb.GetProgress(),
		CreatedAt: pb.GetCreatedAt().AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: pb.GetUpdatedAt().AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapChecklistsFromProto(pbs []*boardpb.Checklist) []checklistResponse {
	result := make([]checklistResponse, 0, len(pbs))
	for _, pb := range pbs {
		result = append(result, mapChecklistFromProto(pb))
	}
	return result
}

func mapChecklistItemFromProto(pb *boardpb.ChecklistItem) checklistItemResponse {
	return checklistItemResponse{
		ID:          pb.GetId(),
		ChecklistID: pb.GetChecklistId(),
		Title:       pb.GetTitle(),
		IsChecked:   pb.GetIsChecked(),
		Position:    pb.GetPosition(),
		CreatedAt:   pb.GetCreatedAt().AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   pb.GetUpdatedAt().AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func parseIntQueryParam(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return n
}
