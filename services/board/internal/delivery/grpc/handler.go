package grpc

import (
	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
)

// BoardServiceServer реализует gRPC сервер Board Service.
// Зависимости сгруппированы по доменным областям в суб-хендлеры.
type BoardServiceServer struct {
	boardpb.UnimplementedBoardServiceServer

	boards       BoardCoreHandler
	columns      ColumnHandler
	cards        CardHandler
	members      MemberHandler
	attachments  AttachmentHandler
	labels       LabelHandler
	cardLinks    CardLinkHandler
	checklists   ChecklistHandler
	customFields CustomFieldHandler
	automations  AutomationHandler
}

// NewBoardServiceServer создает gRPC сервер с доменными суб-хендлерами
func NewBoardServiceServer(
	boards BoardCoreHandler,
	columns ColumnHandler,
	cards CardHandler,
	members MemberHandler,
	attachments AttachmentHandler,
	labels LabelHandler,
	cardLinks CardLinkHandler,
	checklists ChecklistHandler,
	customFields CustomFieldHandler,
	automations AutomationHandler,
) *BoardServiceServer {
	return &BoardServiceServer{
		boards:       boards,
		columns:      columns,
		cards:        cards,
		members:      members,
		attachments:  attachments,
		labels:       labels,
		cardLinks:    cardLinks,
		checklists:   checklists,
		customFields: customFields,
		automations:  automations,
	}
}

// stringPtr возвращает указатель на строку
func stringPtr(s string) *string {
	return &s
}
