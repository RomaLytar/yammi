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
	createCard *usecase.CreateCardUseCase
	getCard    *usecase.GetCardUseCase
	getCards   *usecase.GetCardsUseCase
	updateCard *usecase.UpdateCardUseCase
	moveCard   *usecase.MoveCardUseCase
	deleteCard *usecase.DeleteCardUseCase

	// Member Use Cases
	addMember    *usecase.AddMemberUseCase
	removeMember *usecase.RemoveMemberUseCase
	listMembers  *usecase.ListMembersUseCase
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
	addMember *usecase.AddMemberUseCase,
	removeMember *usecase.RemoveMemberUseCase,
	listMembers *usecase.ListMembersUseCase,
) *BoardServiceServer {
	return &BoardServiceServer{
		createBoard:    createBoard,
		getBoard:       getBoard,
		listBoards:     listBoards,
		updateBoard:    updateBoard,
		deleteBoard:    deleteBoard,
		addColumn:      addColumn,
		getColumns:     getColumns,
		updateColumn:   updateColumn,
		deleteColumn:   deleteColumn,
		reorderColumns: reorderColumns,
		createCard:     createCard,
		getCard:        getCard,
		getCards:       getCards,
		updateCard:     updateCard,
		moveCard:       moveCard,
		deleteCard:     deleteCard,
		addMember:      addMember,
		removeMember:   removeMember,
		listMembers:    listMembers,
	}
}

// stringPtr возвращает указатель на строку
func stringPtr(s string) *string {
	return &s
}
