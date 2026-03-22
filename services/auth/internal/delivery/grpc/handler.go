package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RomaLytar/yammi/services/auth/internal/domain"
	"github.com/RomaLytar/yammi/services/auth/internal/usecase"
	authpb "github.com/RomaLytar/yammi/services/auth/api/proto/v1"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	uc *usecase.AuthUseCase
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "email, password and name are required")
	}

	userID, accessToken, refreshToken, err := h.uc.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetName())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &authpb.RegisterResponse{
		UserId:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	userID, accessToken, refreshToken, err := h.uc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &authpb.LoginResponse{
		UserId:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	accessToken, refreshToken, err := h.uc.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &authpb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *AuthHandler) RevokeToken(ctx context.Context, req *authpb.RevokeTokenRequest) (*authpb.RevokeTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	if err := h.uc.RevokeToken(ctx, req.GetRefreshToken()); err != nil {
		return nil, mapDomainError(err)
	}

	return &authpb.RevokeTokenResponse{}, nil
}

func (h *AuthHandler) GetPublicKey(ctx context.Context, req *authpb.GetPublicKeyRequest) (*authpb.GetPublicKeyResponse, error) {
	pem, algorithm := h.uc.GetPublicKey()
	return &authpb.GetPublicKeyResponse{
		PublicKeyPem: pem,
		Algorithm:    algorithm,
	}, nil
}

func (h *AuthHandler) DeleteUser(ctx context.Context, req *authpb.DeleteUserRequest) (*authpb.DeleteUserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if err := h.uc.DeleteUser(ctx, req.GetUserId()); err != nil {
		return nil, mapDomainError(err)
	}

	return &authpb.DeleteUserResponse{}, nil
}

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrTokenNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidPassword):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrTokenRevoked), errors.Is(err, domain.ErrTokenExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrEmptyEmail), errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrEmptyPassword), errors.Is(err, domain.ErrEmptyName),
		errors.Is(err, domain.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
