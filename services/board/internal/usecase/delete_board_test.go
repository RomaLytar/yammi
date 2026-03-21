package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteBoard_Success_Single(t *testing.T) {
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	boardRepo.On("BatchDelete", mock.Anything, []string{"board-123"}).
		Return(nil)
	publisher.On("PublishBoardDeleted", mock.Anything, mock.Anything).Return(nil).Maybe()

	useCase := NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	err := useCase.Execute(context.Background(), []string{"board-123"}, "user-123")

	assert.NoError(t, err)

	boardRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestDeleteBoard_Success_Batch(t *testing.T) {
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	boardIDs := []string{"board-1", "board-2", "board-3"}

	memberRepo.On("IsMember", mock.Anything, "board-1", "user-123").
		Return(true, domain.RoleOwner, nil)
	memberRepo.On("IsMember", mock.Anything, "board-2", "user-123").
		Return(true, domain.RoleOwner, nil)
	memberRepo.On("IsMember", mock.Anything, "board-3", "user-123").
		Return(true, domain.RoleOwner, nil)
	boardRepo.On("BatchDelete", mock.Anything, boardIDs).
		Return(nil)
	publisher.On("PublishBoardDeleted", mock.Anything, mock.Anything).Return(nil).Maybe()

	useCase := NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	err := useCase.Execute(context.Background(), boardIDs, "user-123")

	assert.NoError(t, err)

	boardRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestDeleteBoard_AccessDenied_NotOwner(t *testing.T) {
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	// Пользователь является участником, но не owner'ом
	memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").
		Return(true, domain.RoleMember, nil)

	useCase := NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	err := useCase.Execute(context.Background(), []string{"board-123"}, "user-456")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)

	boardRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestDeleteBoard_AccessDenied_NotMember(t *testing.T) {
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	// Пользователь вообще не является участником доски
	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	useCase := NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	err := useCase.Execute(context.Background(), []string{"board-123"}, "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)

	boardRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestDeleteBoard_BatchPartialOwnership(t *testing.T) {
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	// Пользователь — owner доски A, но только member доски B
	memberRepo.On("IsMember", mock.Anything, "board-a", "user-123").
		Return(true, domain.RoleOwner, nil)
	memberRepo.On("IsMember", mock.Anything, "board-b", "user-123").
		Return(true, domain.RoleMember, nil)

	useCase := NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	err := useCase.Execute(context.Background(), []string{"board-a", "board-b"}, "user-123")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)

	// BatchDelete не должен быть вызван — проверка ownership упала раньше
	boardRepo.AssertNotCalled(t, "BatchDelete", mock.Anything, mock.Anything)
	boardRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestDeleteBoard_IsMemberError(t *testing.T) {
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(false, domain.Role(""), errors.New("database error"))

	useCase := NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	err := useCase.Execute(context.Background(), []string{"board-123"}, "user-123")

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())

	boardRepo.AssertNotCalled(t, "BatchDelete", mock.Anything, mock.Anything)
	boardRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestDeleteBoard_BatchDeleteError(t *testing.T) {
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	boardRepo.On("BatchDelete", mock.Anything, []string{"board-123"}).
		Return(errors.New("database error"))

	useCase := NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	err := useCase.Execute(context.Background(), []string{"board-123"}, "user-123")

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())

	boardRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}
