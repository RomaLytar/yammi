package grpc

import (
	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// BoardServiceServer реализует gRPC сервер Board Service
type BoardServiceServer struct {
	boardpb.UnimplementedBoardServiceServer

	// Board Use Cases
	createBoard *usecase.CreateBoardUseCase
	getBoard    *usecase.GetBoardUseCase
	listBoards  *usecase.ListBoardsUseCase
	updateBoard *usecase.UpdateBoardUseCase
	deleteBoard *usecase.DeleteBoardUseCase

	// Column Use Cases
	addColumn      *usecase.AddColumnUseCase
	getColumns     *usecase.GetColumnsUseCase
	updateColumn   *usecase.UpdateColumnUseCase
	deleteColumn   *usecase.DeleteColumnUseCase
	reorderColumns *usecase.ReorderColumnsUseCase

	// Card Use Cases
	createCard   *usecase.CreateCardUseCase
	getCard      *usecase.GetCardUseCase
	getCards     *usecase.GetCardsUseCase
	updateCard   *usecase.UpdateCardUseCase
	moveCard     *usecase.MoveCardUseCase
	deleteCard   *usecase.DeleteCardUseCase
	assignCard       *usecase.AssignCardUseCase
	unassignCard     *usecase.UnassignCardUseCase
	listCardActivity *usecase.ListCardActivityUseCase

	// Member Use Cases
	addMember    *usecase.AddMemberUseCase
	removeMember *usecase.RemoveMemberUseCase
	listMembers  *usecase.ListMembersUseCase

	// Card Repository (для CountByBoard в GetBoard)
	cardRepo usecase.CardRepository

	// Attachment Use Cases
	uploadAttachment *usecase.UploadAttachmentUseCase
	confirmUpload    *usecase.ConfirmUploadUseCase
	getDownloadURL   *usecase.GetDownloadURLUseCase
	listAttachments  *usecase.ListAttachmentsUseCase
	deleteAttachment *usecase.DeleteAttachmentUseCase

	// Label Use Cases
	createLabel        *usecase.CreateLabelUseCase
	listLabels         *usecase.ListLabelsUseCase
	updateLabel        *usecase.UpdateLabelUseCase
	deleteLabel        *usecase.DeleteLabelUseCase
	addLabelToCard     *usecase.AddLabelToCardUseCase
	removeLabelFromCard *usecase.RemoveLabelFromCardUseCase
	getCardLabels      *usecase.GetCardLabelsUseCase

	// Card Link Use Cases
	linkCards       *usecase.LinkCardsUseCase
	unlinkCards     *usecase.UnlinkCardsUseCase
	getCardChildren *usecase.GetCardChildrenUseCase
	getCardParents  *usecase.GetCardParentsUseCase

	// Checklist Use Cases
	createChecklist     *usecase.CreateChecklistUseCase
	getChecklists       *usecase.GetChecklistsUseCase
	updateChecklist     *usecase.UpdateChecklistUseCase
	deleteChecklist     *usecase.DeleteChecklistUseCase
	createChecklistItem *usecase.CreateChecklistItemUseCase
	updateChecklistItem *usecase.UpdateChecklistItemUseCase
	deleteChecklistItem *usecase.DeleteChecklistItemUseCase
	toggleChecklistItem *usecase.ToggleChecklistItemUseCase

	// Custom Field Use Cases
	createCustomField    *usecase.CreateCustomFieldUseCase
	listCustomFields     *usecase.ListCustomFieldsUseCase
	updateCustomField    *usecase.UpdateCustomFieldUseCase
	deleteCustomField    *usecase.DeleteCustomFieldUseCase
	setCustomFieldValue  *usecase.SetCustomFieldValueUseCase
	getCardCustomFields  *usecase.GetCardCustomFieldsUseCase

	// Automation Rule Use Cases
	createAutomationRule *usecase.CreateAutomationRuleUseCase
	listAutomationRules  *usecase.ListAutomationRulesUseCase
	updateAutomationRule *usecase.UpdateAutomationRuleUseCase
	deleteAutomationRule *usecase.DeleteAutomationRuleUseCase
	getAutomationHistory *usecase.GetAutomationHistoryUseCase
}

// NewBoardServiceServer создает новый gRPC сервер с внедренными use cases
func NewBoardServiceServer(
	createBoard *usecase.CreateBoardUseCase,
	getBoard *usecase.GetBoardUseCase,
	listBoards *usecase.ListBoardsUseCase,
	updateBoard *usecase.UpdateBoardUseCase,
	deleteBoard *usecase.DeleteBoardUseCase,
	addColumn *usecase.AddColumnUseCase,
	getColumns *usecase.GetColumnsUseCase,
	updateColumn *usecase.UpdateColumnUseCase,
	deleteColumn *usecase.DeleteColumnUseCase,
	reorderColumns *usecase.ReorderColumnsUseCase,
	createCard *usecase.CreateCardUseCase,
	getCard *usecase.GetCardUseCase,
	getCards *usecase.GetCardsUseCase,
	updateCard *usecase.UpdateCardUseCase,
	moveCard *usecase.MoveCardUseCase,
	deleteCard *usecase.DeleteCardUseCase,
	assignCard *usecase.AssignCardUseCase,
	unassignCard *usecase.UnassignCardUseCase,
	listCardActivity *usecase.ListCardActivityUseCase,
	addMember *usecase.AddMemberUseCase,
	removeMember *usecase.RemoveMemberUseCase,
	listMembers *usecase.ListMembersUseCase,
	cardRepo usecase.CardRepository,
	uploadAttachment *usecase.UploadAttachmentUseCase,
	confirmUpload *usecase.ConfirmUploadUseCase,
	getDownloadURL *usecase.GetDownloadURLUseCase,
	listAttachments *usecase.ListAttachmentsUseCase,
	deleteAttachment *usecase.DeleteAttachmentUseCase,
	createLabel *usecase.CreateLabelUseCase,
	listLabelsUC *usecase.ListLabelsUseCase,
	updateLabel *usecase.UpdateLabelUseCase,
	deleteLabel *usecase.DeleteLabelUseCase,
	addLabelToCard *usecase.AddLabelToCardUseCase,
	removeLabelFromCard *usecase.RemoveLabelFromCardUseCase,
	getCardLabels *usecase.GetCardLabelsUseCase,
	linkCards *usecase.LinkCardsUseCase,
	unlinkCards *usecase.UnlinkCardsUseCase,
	getCardChildren *usecase.GetCardChildrenUseCase,
	getCardParents *usecase.GetCardParentsUseCase,
	createChecklist *usecase.CreateChecklistUseCase,
	getChecklists *usecase.GetChecklistsUseCase,
	updateChecklist *usecase.UpdateChecklistUseCase,
	deleteChecklist *usecase.DeleteChecklistUseCase,
	createChecklistItem *usecase.CreateChecklistItemUseCase,
	updateChecklistItem *usecase.UpdateChecklistItemUseCase,
	deleteChecklistItem *usecase.DeleteChecklistItemUseCase,
	toggleChecklistItem *usecase.ToggleChecklistItemUseCase,
	createCustomField *usecase.CreateCustomFieldUseCase,
	listCustomFieldsUC *usecase.ListCustomFieldsUseCase,
	updateCustomField *usecase.UpdateCustomFieldUseCase,
	deleteCustomField *usecase.DeleteCustomFieldUseCase,
	setCustomFieldValue *usecase.SetCustomFieldValueUseCase,
	getCardCustomFields *usecase.GetCardCustomFieldsUseCase,
	createAutomationRule *usecase.CreateAutomationRuleUseCase,
	listAutomationRules *usecase.ListAutomationRulesUseCase,
	updateAutomationRule *usecase.UpdateAutomationRuleUseCase,
	deleteAutomationRule *usecase.DeleteAutomationRuleUseCase,
	getAutomationHistory *usecase.GetAutomationHistoryUseCase,
) *BoardServiceServer {
	return &BoardServiceServer{
		createBoard:         createBoard,
		getBoard:            getBoard,
		listBoards:          listBoards,
		updateBoard:         updateBoard,
		deleteBoard:         deleteBoard,
		addColumn:           addColumn,
		getColumns:          getColumns,
		updateColumn:        updateColumn,
		deleteColumn:        deleteColumn,
		reorderColumns:      reorderColumns,
		createCard:          createCard,
		getCard:             getCard,
		getCards:            getCards,
		updateCard:          updateCard,
		moveCard:            moveCard,
		deleteCard:          deleteCard,
		assignCard:          assignCard,
		unassignCard:        unassignCard,
		listCardActivity:    listCardActivity,
		addMember:           addMember,
		removeMember:        removeMember,
		listMembers:         listMembers,
		cardRepo:            cardRepo,
		uploadAttachment:    uploadAttachment,
		confirmUpload:       confirmUpload,
		getDownloadURL:      getDownloadURL,
		listAttachments:     listAttachments,
		deleteAttachment:    deleteAttachment,
		createLabel:         createLabel,
		listLabels:          listLabelsUC,
		updateLabel:         updateLabel,
		deleteLabel:         deleteLabel,
		addLabelToCard:      addLabelToCard,
		removeLabelFromCard: removeLabelFromCard,
		getCardLabels:       getCardLabels,
		linkCards:            linkCards,
		unlinkCards:          unlinkCards,
		getCardChildren:     getCardChildren,
		getCardParents:      getCardParents,
		createChecklist:     createChecklist,
		getChecklists:       getChecklists,
		updateChecklist:     updateChecklist,
		deleteChecklist:     deleteChecklist,
		createChecklistItem: createChecklistItem,
		updateChecklistItem: updateChecklistItem,
		deleteChecklistItem: deleteChecklistItem,
		toggleChecklistItem:  toggleChecklistItem,
		createCustomField:    createCustomField,
		listCustomFields:     listCustomFieldsUC,
		updateCustomField:    updateCustomField,
		deleteCustomField:    deleteCustomField,
		setCustomFieldValue:  setCustomFieldValue,
		getCardCustomFields:  getCardCustomFields,
		createAutomationRule: createAutomationRule,
		listAutomationRules:  listAutomationRules,
		updateAutomationRule: updateAutomationRule,
		deleteAutomationRule: deleteAutomationRule,
		getAutomationHistory: getAutomationHistory,
	}
}

// stringPtr возвращает указатель на строку
func stringPtr(s string) *string {
	return &s
}
