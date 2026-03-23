package usecase

import "context"

// EventPublisher публикует события в NATS
type EventPublisher interface {
	PublishBoardCreated(ctx context.Context, event BoardCreated) error
	PublishBoardUpdated(ctx context.Context, event BoardUpdated) error
	PublishBoardDeleted(ctx context.Context, event BoardDeleted) error
	PublishColumnCreated(ctx context.Context, event ColumnAdded) error
	PublishColumnUpdated(ctx context.Context, event ColumnUpdated) error
	PublishColumnDeleted(ctx context.Context, event ColumnDeleted) error
	PublishColumnsReordered(ctx context.Context, event ColumnsReordered) error
	PublishCardCreated(ctx context.Context, event CardCreated) error
	PublishCardUpdated(ctx context.Context, event CardUpdated) error
	PublishCardMoved(ctx context.Context, event CardMoved) error
	PublishCardDeleted(ctx context.Context, event CardDeleted) error
	PublishMemberAdded(ctx context.Context, event MemberAdded) error
	PublishMemberRemoved(ctx context.Context, event MemberRemoved) error
	PublishCardAssigned(ctx context.Context, event CardAssigned) error
	PublishCardUnassigned(ctx context.Context, event CardUnassigned) error
	PublishAttachmentUploaded(ctx context.Context, event AttachmentUploaded) error
	PublishAttachmentDeleted(ctx context.Context, event AttachmentDeleted) error
	PublishLabelCreated(ctx context.Context, event LabelCreated) error
	PublishLabelUpdated(ctx context.Context, event LabelUpdated) error
	PublishLabelDeleted(ctx context.Context, event LabelDeleted) error
	PublishCardLabelAdded(ctx context.Context, event CardLabelAdded) error
	PublishCardLabelRemoved(ctx context.Context, event CardLabelRemoved) error
	PublishCardLinked(ctx context.Context, event CardLinked) error
	PublishCardUnlinked(ctx context.Context, event CardUnlinked) error
	PublishChecklistCreated(ctx context.Context, event ChecklistCreated) error
	PublishChecklistUpdated(ctx context.Context, event ChecklistUpdated) error
	PublishChecklistDeleted(ctx context.Context, event ChecklistDeleted) error
	PublishChecklistItemToggled(ctx context.Context, event ChecklistItemToggled) error
	PublishCustomFieldCreated(ctx context.Context, event CustomFieldCreated) error
	PublishCustomFieldUpdated(ctx context.Context, event CustomFieldUpdated) error
	PublishCustomFieldDeleted(ctx context.Context, event CustomFieldDeleted) error
	PublishCustomFieldValueSet(ctx context.Context, event CustomFieldValueSet) error
	PublishAutomationRuleCreated(ctx context.Context, event AutomationRuleCreated) error
	PublishAutomationRuleUpdated(ctx context.Context, event AutomationRuleUpdated) error
	PublishAutomationRuleDeleted(ctx context.Context, event AutomationRuleDeleted) error
	PublishAutomationExecuted(ctx context.Context, event AutomationExecuted) error
}
