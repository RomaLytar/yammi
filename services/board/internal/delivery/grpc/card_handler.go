package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
)

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

	// CreateCardRequest не имеет assignee_id в proto (line 153-160), передаем nil
	card, err := s.createCard.Execute(ctx, req.GetColumnId(), req.GetBoardId(), req.GetUserId(), req.GetTitle(), req.GetDescription(), req.GetPosition(), nil)
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

	card, err := s.getCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
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

	cards, err := s.getCards.Execute(ctx, req.GetColumnId(), req.GetBoardId(), req.GetUserId())
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

	card, err := s.updateCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId(), req.GetTitle(), req.GetDescription(), assigneeID, int(req.GetVersion()))
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
	card, err := s.moveCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetFromColumnId(), req.GetToColumnId(), req.GetUserId(), req.GetPosition())
	if err != nil {
		return nil, mapDomainError(err)
	}

	// Доступ проверен в moveCard — загружаем карточки без повторного IsMember
	cardsInColumn, err := s.getCards.ExecuteAuthorized(ctx, req.GetToColumnId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.MoveCardResponse{
		Card:          mapCardToProto(card, req.GetBoardId()),
		CardsInColumn: mapCardsToProto(cardsInColumn, req.GetBoardId()),
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

	err := s.deleteCard.Execute(ctx, req.GetCardIds(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}
