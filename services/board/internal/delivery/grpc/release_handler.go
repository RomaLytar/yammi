package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// ReleaseHandler группирует зависимости для операций с релизами
type ReleaseHandler struct {
	create     *usecase.CreateReleaseUseCase
	get        *usecase.GetReleaseUseCase
	list       *usecase.ListReleasesUseCase
	update     *usecase.UpdateReleaseUseCase
	delete_    *usecase.DeleteReleaseUseCase
	start      *usecase.StartReleaseUseCase
	complete   *usecase.CompleteReleaseUseCase
	getActive  *usecase.GetActiveReleaseUseCase
	assignCard *usecase.AssignCardToReleaseUseCase
	removeCard *usecase.RemoveCardFromReleaseUseCase
	getBacklog *usecase.GetBacklogUseCase
	getCards   *usecase.GetReleaseCardsUseCase
}

func NewReleaseHandler(
	create *usecase.CreateReleaseUseCase,
	get *usecase.GetReleaseUseCase,
	list *usecase.ListReleasesUseCase,
	update *usecase.UpdateReleaseUseCase,
	delete_ *usecase.DeleteReleaseUseCase,
	start *usecase.StartReleaseUseCase,
	complete *usecase.CompleteReleaseUseCase,
	getActive *usecase.GetActiveReleaseUseCase,
	assignCard *usecase.AssignCardToReleaseUseCase,
	removeCard *usecase.RemoveCardFromReleaseUseCase,
	getBacklog *usecase.GetBacklogUseCase,
	getCards *usecase.GetReleaseCardsUseCase,
) ReleaseHandler {
	return ReleaseHandler{
		create: create, get: get, list: list, update: update,
		delete_: delete_, start: start, complete: complete,
		getActive: getActive, assignCard: assignCard, removeCard: removeCard,
		getBacklog: getBacklog, getCards: getCards,
	}
}

// CreateRelease создает новый релиз
func (s *BoardServiceServer) CreateRelease(ctx context.Context, req *boardpb.CreateReleaseRequest) (*boardpb.CreateReleaseResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	var startDate, endDate *time.Time
	if req.GetStartDate() != nil {
		t := req.GetStartDate().AsTime()
		startDate = &t
	}
	if req.GetEndDate() != nil {
		t := req.GetEndDate().AsTime()
		endDate = &t
	}

	release, err := s.releases.create.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetName(), req.GetDescription(), startDate, endDate)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateReleaseResponse{
		Release: mapReleaseToProto(release),
	}, nil
}

// GetRelease возвращает релиз по ID
func (s *BoardServiceServer) GetRelease(ctx context.Context, req *boardpb.GetReleaseRequest) (*boardpb.GetReleaseResponse, error) {
	if req.GetReleaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "release_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	release, err := s.releases.get.Execute(ctx, req.GetReleaseId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetReleaseResponse{
		Release: mapReleaseToProto(release),
	}, nil
}

// ListReleases возвращает все релизы доски
func (s *BoardServiceServer) ListReleases(ctx context.Context, req *boardpb.ListReleasesRequest) (*boardpb.ListReleasesResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	releases, err := s.releases.list.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListReleasesResponse{
		Releases: mapReleasesToProto(releases),
	}, nil
}

// UpdateRelease обновляет релиз
func (s *BoardServiceServer) UpdateRelease(ctx context.Context, req *boardpb.UpdateReleaseRequest) (*boardpb.UpdateReleaseResponse, error) {
	if req.GetReleaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "release_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	var startDate, endDate *time.Time
	if req.GetStartDate() != nil {
		t := req.GetStartDate().AsTime()
		startDate = &t
	}
	if req.GetEndDate() != nil {
		t := req.GetEndDate().AsTime()
		endDate = &t
	}

	release, err := s.releases.update.Execute(ctx, req.GetReleaseId(), req.GetBoardId(), req.GetUserId(), req.GetName(), req.GetDescription(), startDate, endDate)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateReleaseResponse{
		Release: mapReleaseToProto(release),
	}, nil
}

// DeleteRelease удаляет релиз
func (s *BoardServiceServer) DeleteRelease(ctx context.Context, req *boardpb.DeleteReleaseRequest) (*emptypb.Empty, error) {
	if req.GetReleaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "release_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.releases.delete_.Execute(ctx, req.GetReleaseId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// StartRelease переводит релиз из draft в active
func (s *BoardServiceServer) StartRelease(ctx context.Context, req *boardpb.StartReleaseRequest) (*boardpb.StartReleaseResponse, error) {
	if req.GetReleaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "release_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	release, err := s.releases.start.Execute(ctx, req.GetReleaseId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.StartReleaseResponse{
		Release: mapReleaseToProto(release),
	}, nil
}

// CompleteRelease переводит релиз из active в completed
func (s *BoardServiceServer) CompleteRelease(ctx context.Context, req *boardpb.CompleteReleaseRequest) (*boardpb.CompleteReleaseResponse, error) {
	if req.GetReleaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "release_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	release, movedToBacklog, err := s.releases.complete.Execute(ctx, req.GetReleaseId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CompleteReleaseResponse{
		Release:             mapReleaseToProto(release),
		CardsMovedToBacklog: safeInt32(movedToBacklog),
	}, nil
}

// GetActiveRelease возвращает активный релиз доски (если есть)
func (s *BoardServiceServer) GetActiveRelease(ctx context.Context, req *boardpb.GetActiveReleaseRequest) (*boardpb.GetActiveReleaseResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	release, err := s.releases.getActive.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	resp := &boardpb.GetActiveReleaseResponse{}
	if release != nil {
		resp.Release = mapReleaseToProto(release)
	}
	return resp, nil
}

// AssignCardToRelease назначает карточку в релиз
func (s *BoardServiceServer) AssignCardToRelease(ctx context.Context, req *boardpb.AssignCardToReleaseRequest) (*emptypb.Empty, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetReleaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "release_id is required")
	}
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}

	err := s.releases.assignCard.Execute(ctx, req.GetCardId(), req.GetReleaseId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// RemoveCardFromRelease снимает карточку с релиза
func (s *BoardServiceServer) RemoveCardFromRelease(ctx context.Context, req *boardpb.RemoveCardFromReleaseRequest) (*emptypb.Empty, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}

	err := s.releases.removeCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetBacklog возвращает карточки без релиза (бэклог)
func (s *BoardServiceServer) GetBacklog(ctx context.Context, req *boardpb.GetBacklogRequest) (*boardpb.GetBacklogResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	cards, err := s.releases.getBacklog.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetBacklogResponse{
		Cards: mapCardsToProto(cards, req.GetBoardId()),
	}, nil
}

// GetReleaseCards возвращает карточки релиза
func (s *BoardServiceServer) GetReleaseCards(ctx context.Context, req *boardpb.GetReleaseCardsRequest) (*boardpb.GetReleaseCardsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetReleaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "release_id is required")
	}

	cards, err := s.releases.getCards.Execute(ctx, req.GetBoardId(), req.GetReleaseId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetReleaseCardsResponse{
		Cards: mapCardsToProto(cards, req.GetBoardId()),
	}, nil
}
