package domain

import "errors"

// Board errors
var (
	ErrBoardNotFound    = errors.New("board not found")
	ErrEmptyTitle       = errors.New("board title cannot be empty")
	ErrEmptyOwnerID     = errors.New("owner ID cannot be empty")
	ErrAccessDenied     = errors.New("access denied")
	ErrNotOwner         = errors.New("only owner can perform this action")
	ErrInvalidVersion   = errors.New("invalid version for optimistic locking")
)

// Column errors
var (
	ErrColumnNotFound = errors.New("column not found")
	ErrEmptyColumnTitle = errors.New("column title cannot be empty")
	ErrInvalidPosition = errors.New("invalid position")
)

// Card errors
var (
	ErrCardNotFound      = errors.New("card not found")
	ErrEmptyCardTitle    = errors.New("card title cannot be empty")
	ErrInvalidLexorank   = errors.New("invalid lexorank position")
	ErrCardNotInColumn   = errors.New("card does not belong to this column")
	ErrAssigneeNotMember = errors.New("assignee is not a board member")
	ErrInvalidPriority   = errors.New("invalid priority")
	ErrInvalidTaskType   = errors.New("invalid task type")
)

// Member errors
var (
	ErrMemberNotFound    = errors.New("member not found")
	ErrMemberExists      = errors.New("member already exists")
	ErrCannotRemoveOwner = errors.New("cannot remove owner from board")
	ErrInvalidRole       = errors.New("invalid role")
)

// Attachment errors
var (
	ErrAttachmentNotFound    = errors.New("attachment not found")
	ErrFileTooLarge          = errors.New("file size exceeds maximum allowed (50 MB)")
	ErrMaxAttachmentsReached = errors.New("maximum number of attachments per card reached")
	ErrEmptyFileName         = errors.New("file name cannot be empty")
	ErrFileSizeMismatch      = errors.New("uploaded file size does not match declared size")
	ErrFileTypeMismatch      = errors.New("uploaded file content type does not match declared type")
)

// Label errors
var (
	ErrLabelNotFound      = errors.New("label not found")
	ErrLabelExists        = errors.New("label with this name already exists")
	ErrEmptyLabelName     = errors.New("label name cannot be empty")
	ErrInvalidColor       = errors.New("invalid color format")
	ErrMaxLabelsReached   = errors.New("maximum labels per board reached")
	ErrLabelAlreadyOnCard = errors.New("label already assigned to card")
)

// Checklist errors
var (
	ErrChecklistNotFound     = errors.New("checklist not found")
	ErrChecklistItemNotFound = errors.New("checklist item not found")
	ErrEmptyChecklistTitle   = errors.New("checklist title cannot be empty")
	ErrEmptyItemTitle        = errors.New("checklist item title cannot be empty")
)

// Card Link errors
var (
	ErrCardLinkNotFound = errors.New("card link not found")
	ErrSelfLink         = errors.New("card cannot link to itself")
	ErrLinkAlreadyExists = errors.New("link already exists between these cards")
	ErrInvalidLinkType  = errors.New("invalid link type")
)

// Custom Field errors
var (
	ErrCustomFieldNotFound      = errors.New("custom field not found")
	ErrCustomFieldValueNotFound = errors.New("custom field value not found")
	ErrCustomFieldExists        = errors.New("custom field with this name already exists")
	ErrEmptyFieldName           = errors.New("field name cannot be empty")
	ErrInvalidFieldType         = errors.New("invalid field type")
	ErrInvalidFieldValue        = errors.New("invalid field value for field type")
	ErrMaxCustomFieldsReached   = errors.New("maximum custom fields per board reached")
)

// Automation errors
var (
	ErrAutomationRuleNotFound = errors.New("automation rule not found")
	ErrEmptyRuleName          = errors.New("rule name cannot be empty")
	ErrInvalidTriggerType     = errors.New("invalid trigger type")
	ErrInvalidActionType      = errors.New("invalid action type")
	ErrMaxRulesReached        = errors.New("maximum automation rules per board reached")
)

// User Label errors
var (
	ErrUserLabelNotFound    = errors.New("user label not found")
	ErrUserLabelExists      = errors.New("user label with this name already exists")
	ErrMaxUserLabelsReached = errors.New("maximum user labels reached")
)

// Template errors
var (
	ErrTemplateNotFound    = errors.New("template not found")
	ErrEmptyTemplateName   = errors.New("template name cannot be empty")
	ErrMaxTemplatesReached = errors.New("maximum templates per board reached")
)

// Release errors
var (
	ErrReleaseNotFound        = errors.New("release not found")
	ErrEmptyReleaseName       = errors.New("release name cannot be empty")
	ErrReleaseCompleted       = errors.New("completed release cannot be modified")
	ErrReleaseNotDraft        = errors.New("release must be in draft status")
	ErrReleaseNotActive       = errors.New("release must be in active status")
	ErrActiveReleaseExists    = errors.New("board already has an active release")
	ErrMaxReleasesReached     = errors.New("maximum releases per board reached")
	ErrCardInCompletedRelease = errors.New("cannot modify card in a completed release")
	ErrInvalidSprintDuration  = errors.New("sprint duration must be at least 7 days")
)

// Activity errors
var (
	ErrEmptyActorID      = errors.New("actor ID cannot be empty")
	ErrInvalidActivityType = errors.New("invalid activity type")
)
