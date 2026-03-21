package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "github.com/romanlovesweed/yammi/services/user/api/proto/v1"
	"github.com/romanlovesweed/yammi/services/user/internal/domain"
	"github.com/romanlovesweed/yammi/services/user/internal/usecase"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	uc *usecase.UserUseCase
}

func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.GetProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := h.uc.GetProfile(ctx, req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &userpb.GetProfileResponse{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarUrl: user.AvatarURL,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UpdateProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := h.uc.UpdateProfile(ctx, req.GetUserId(), req.GetName(), req.GetAvatarUrl(), req.GetBio())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &userpb.UpdateProfileResponse{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarUrl: user.AvatarURL,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *UserHandler) SearchByEmail(ctx context.Context, req *userpb.SearchByEmailRequest) (*userpb.SearchByEmailResponse, error) {
	if req.GetQuery() == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	users, err := h.uc.SearchByEmail(ctx, req.GetQuery(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	result := make([]*userpb.UserInfo, 0, len(users))
	for _, u := range users {
		result = append(result, &userpb.UserInfo{
			Id:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			AvatarUrl: u.AvatarURL,
		})
	}

	return &userpb.SearchByEmailResponse{Users: result}, nil
}

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrEmptyName):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
