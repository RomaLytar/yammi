package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

// AddMember добавляет участника в доску (только owner)
func (s *BoardServiceServer) AddMember(ctx context.Context, req *boardpb.AddMemberRequest) (*boardpb.AddMemberResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetMemberUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "member_user_id is required")
	}
	if req.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}

	role := domain.Role(req.GetRole())
	if !role.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	err := s.addMember.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetMemberUserId(), role)
	if err != nil {
		return nil, mapDomainError(err)
	}

	// AddMemberUseCase не возвращает member, загружаем его отдельно
	members, err := s.listMembers.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	// Находим добавленного member
	var addedMember *domain.Member
	for _, m := range members {
		if m.UserID == req.GetMemberUserId() {
			addedMember = m
			break
		}
	}

	if addedMember == nil {
		return nil, status.Error(codes.Internal, "member was added but not found")
	}

	return &boardpb.AddMemberResponse{
		Member: mapMemberToProto(addedMember),
	}, nil
}

// RemoveMember удаляет участника из доски (только owner)
func (s *BoardServiceServer) RemoveMember(ctx context.Context, req *boardpb.RemoveMemberRequest) (*emptypb.Empty, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetMemberUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "member_user_id is required")
	}

	err := s.removeMember.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetMemberUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// ListMembers возвращает список участников доски
func (s *BoardServiceServer) ListMembers(ctx context.Context, req *boardpb.ListMembersRequest) (*boardpb.ListMembersResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	members, err := s.listMembers.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListMembersResponse{
		Members: mapMembersToProto(members),
	}, nil
}

// IsMember проверяет, является ли пользователь участником доски
func (s *BoardServiceServer) IsMember(ctx context.Context, req *boardpb.IsMemberRequest) (*boardpb.IsMemberResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Используем listMembers usecase (у него есть проверка доступа), но для IsMember
	// нам нужен прямой вызов репозитория. Вызываем через memberRepo напрямую — это
	// delivery-level utility для cross-service запросов.
	// Для этого используем listMembers и ищем user_id в результате.
	members, err := s.listMembers.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		// Если access denied — значит не участник
		if errors.Is(err, domain.ErrAccessDenied) {
			return &boardpb.IsMemberResponse{
				IsMember: false,
				Role:     "",
			}, nil
		}
		return nil, mapDomainError(err)
	}

	for _, m := range members {
		if m.UserID == req.GetUserId() {
			return &boardpb.IsMemberResponse{
				IsMember: true,
				Role:     m.Role.String(),
			}, nil
		}
	}

	return &boardpb.IsMemberResponse{
		IsMember: false,
		Role:     "",
	}, nil
}
