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
	}
}

// stringPtr возвращает указатель на строку
func stringPtr(s string) *string {
	return &s
}
