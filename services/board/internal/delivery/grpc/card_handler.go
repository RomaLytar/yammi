package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// CardHandler группирует зависимости для операций с карточками
type CardHandler struct {
	create   *usecase.CreateCardUseCase
	get      *usecase.GetCardUseCase
	getAll   *usecase.GetCardsUseCase
	update   *usecase.UpdateCardUseCase
	move     *usecase.MoveCardUseCase
	delete   *usecase.DeleteCardUseCase
	assign   *usecase.AssignCardUseCase
	unassign *usecase.UnassignCardUseCase
	activity *usecase.ListCardActivityUseCase
	search   *usecase.SearchBoardCardsUseCase
	repo     usecase.CardRepository
}

func NewCardHandler(
	create *usecase.CreateCardUseCase,
	get *usecase.GetCardUseCase,
	getAll *usecase.GetCardsUseCase,
	update *usecase.UpdateCardUseCase,
	move *usecase.MoveCardUseCase,
	delete_ *usecase.DeleteCardUseCase,
	assign *usecase.AssignCardUseCase,
	unassign *usecase.UnassignCardUseCase,
	activity *usecase.ListCardActivityUseCase,
	search *usecase.SearchBoardCardsUseCase,
	repo usecase.CardRepository,
) CardHandler {
	return CardHandler{
		create: create, get: get, getAll: getAll, update: update,
		move: move, delete: delete_, assign: assign, unassign: unassign,
		activity: activity, search: search, repo: repo,
	}
}

// CreateCard создает новую карточку
func (s *BoardServiceServer) CreateCard(ctx context.Context, req *boardpb.CreateCardRequest) (*boardpb.CreateCardResponse, error) {
	if req.GetColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "column_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetPosition() == "" {
		return nil, status.Error(codes.InvalidArgument, "position (lexorank) is required")
	}

	// Преобразуем assignee_id в *string
	var assigneeID *string
	if req.GetAssigneeId() != "" {
		assigneeID = stringPtr(req.GetAssigneeId())
	}

	// Преобразуем due_date из proto timestamp
	var dueDate *time.Time
	if req.GetDueDate() != nil {
		t := req.GetDueDate().AsTime()
		dueDate = &t
	}

	card, err := s.cards.create.Execute(ctx, req.GetColumnId(), req.GetBoardId(), req.GetUserId(), req.GetTitle(), req.GetDescription(), req.GetPosition(), assigneeID, dueDate, domain.Priority(req.GetPriority()), domain.TaskType(req.GetTaskType()))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateCardResponse{
		Card: mapCardToProto(card, req.GetBoardId()),
	}, nil
}

// GetCard возвращает карточку по ID
func (s *BoardServiceServer) GetCard(ctx context.Context, req *boardpb.GetCardRequest) (*boardpb.GetCardResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	card, err := s.cards.get.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetCardResponse{
		Card: mapCardToProto(card, req.GetBoardId()),
	}, nil
}

// GetCards возвращает все карточки колонки
func (s *BoardServiceServer) GetCards(ctx context.Context, req *boardpb.GetCardsRequest) (*boardpb.GetCardsResponse, error) {
	if req.GetColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "column_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	cards, err := s.cards.getAll.Execute(ctx, req.GetColumnId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetCardsResponse{
		Cards: mapCardsToProto(cards, req.GetBoardId()),
	}, nil
}

// UpdateCard обновляет метаданные карточки
func (s *BoardServiceServer) UpdateCard(ctx context.Context, req *boardpb.UpdateCardRequest) (*boardpb.UpdateCardResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Преобразуем assignee_id в *string
	var assigneeID *string
	if req.GetAssigneeId() != "" {
		assigneeID = stringPtr(req.GetAssigneeId())
	}

	// Преобразуем due_date из proto timestamp
	var dueDate *time.Time
	if req.GetDueDate() != nil {
		t := req.GetDueDate().AsTime()
		dueDate = &t
	}

	card, err := s.cards.update.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId(), req.GetTitle(), req.GetDescription(), assigneeID, int(req.GetVersion()), dueDate, domain.Priority(req.GetPriority()), domain.TaskType(req.GetTaskType()))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateCardResponse{
		Card: mapCardToProto(card, req.GetBoardId()),
	}, nil
}

// MoveCard перемещает карточку между колонками (с lexorank позицией)
func (s *BoardServiceServer) MoveCard(ctx context.Context, req *boardpb.MoveCardRequest) (*boardpb.MoveCardResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetFromColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "from_column_id is required")
	}
	if req.GetToColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "to_column_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetPosition() == "" {
		return nil, status.Error(codes.InvalidArgument, "position (lexorank) is required")
	}

	// Передаём lexorank позицию напрямую (вычисляется на фронтенде)
	card, err := s.cards.move.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetFromColumnId(), req.GetToColumnId(), req.GetUserId(), req.GetPosition())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.MoveCardResponse{
		Card: mapCardToProto(card, req.GetBoardId()),
	}, nil
}

// DeleteCard удаляет одну или несколько карточек (batch)
func (s *BoardServiceServer) DeleteCard(ctx context.Context, req *boardpb.DeleteCardRequest) (*emptypb.Empty, error) {
	if len(req.GetCardIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "card_ids is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.cards.delete.Execute(ctx, req.GetCardIds(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// AssignCard назначает карточку на участника доски
func (s *BoardServiceServer) AssignCard(ctx context.Context, req *boardpb.AssignCardRequest) (*boardpb.AssignCardResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetAssigneeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "assignee_id is required")
	}

	card, err := s.cards.assign.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId(), req.GetAssigneeId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.AssignCardResponse{
		Card: mapCardToProto(card, req.GetBoardId()),
	}, nil
}

// UnassignCard снимает назначение с карточки
func (s *BoardServiceServer) UnassignCard(ctx context.Context, req *boardpb.UnassignCardRequest) (*boardpb.UnassignCardResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	card, err := s.cards.unassign.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UnassignCardResponse{
		Card: mapCardToProto(card, req.GetBoardId()),
	}, nil
}

// SearchBoardCards ищет карточки по доске с опциональными фильтрами
func (s *BoardServiceServer) SearchBoardCards(ctx context.Context, req *boardpb.SearchBoardCardsRequest) (*boardpb.SearchBoardCardsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	cards, err := s.cards.search.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetSearch(), req.GetAssigneeId(), req.GetPriority(), req.GetTaskType())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.SearchBoardCardsResponse{
		Cards: mapCardsToProto(cards, req.GetBoardId()),
	}, nil
}
