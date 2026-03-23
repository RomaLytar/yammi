package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
)

// LinkCards создает связь parent->child между карточками
func (s *BoardServiceServer) LinkCards(ctx context.Context, req *boardpb.LinkCardsRequest) (*boardpb.LinkCardsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetParentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "parent_id is required")
	}
	if req.GetChildId() == "" {
		return nil, status.Error(codes.InvalidArgument, "child_id is required")
	}

	link, err := s.linkCards.Execute(ctx, req.GetParentId(), req.GetChildId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.LinkCardsResponse{
		Link: mapCardLinkToProto(link),
	}, nil
}

// UnlinkCards удаляет связь между карточками
func (s *BoardServiceServer) UnlinkCards(ctx context.Context, req *boardpb.UnlinkCardsRequest) (*emptypb.Empty, error) {
	if req.GetLinkId() == "" {
		return nil, status.Error(codes.InvalidArgument, "link_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.unlinkCards.Execute(ctx, req.GetLinkId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetCardChildren возвращает дочерние связи карточки
func (s *BoardServiceServer) GetCardChildren(ctx context.Context, req *boardpb.GetCardChildrenRequest) (*boardpb.GetCardChildrenResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	links, err := s.getCardChildren.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetCardChildrenResponse{
		Links: mapCardLinksToProto(links),
	}, nil
}

// GetCardParents возвращает родительские связи карточки
func (s *BoardServiceServer) GetCardParents(ctx context.Context, req *boardpb.GetCardParentsRequest) (*boardpb.GetCardParentsResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	links, err := s.getCardParents.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetCardParentsResponse{
		Links: mapCardLinksToProto(links),
	}, nil
}
