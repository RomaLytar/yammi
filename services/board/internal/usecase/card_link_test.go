package usecase

import (
	"context"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCardLinkRepository - мок для CardLinkRepository
type MockCardLinkRepository struct {
	mock.Mock
}

func (m *MockCardLinkRepository) Create(ctx context.Context, link *domain.CardLink) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockCardLinkRepository) Delete(ctx context.Context, linkID, boardID string) error {
	args := m.Called(ctx, linkID, boardID)
	return args.Error(0)
}

func (m *MockCardLinkRepository) GetByID(ctx context.Context, linkID, boardID string) (*domain.CardLink, error) {
	args := m.Called(ctx, linkID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CardLink), args.Error(1)
}

func (m *MockCardLinkRepository) ListChildren(ctx context.Context, parentID, boardID string) ([]*domain.CardLink, error) {
	args := m.Called(ctx, parentID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CardLink), args.Error(1)
}

func (m *MockCardLinkRepository) ListParents(ctx context.Context, childID string) ([]*domain.CardLink, error) {
	args := m.Called(ctx, childID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CardLink), args.Error(1)
}

func (m *MockCardLinkRepository) Exists(ctx context.Context, parentID, childID, boardID string) (bool, error) {
	args := m.Called(ctx, parentID, childID, boardID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCardLinkRepository) CreateVerified(ctx context.Context, link *domain.CardLink) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func TestLinkCards_Success(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	cardRepo := new(MockCardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cardLinkRepo.On("CreateVerified", mock.Anything, mock.AnythingOfType("*domain.CardLink")).Return(nil)
	publisher.On("PublishCardLinked", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewLinkCardsUseCase(cardLinkRepo, cardRepo, memberRepo, publisher)
	link, err := uc.Execute(context.Background(), "parent-123", "child-456", "board-123", "user-123")

	assert.NoError(t, err)
	assert.NotNil(t, link)
	assert.Equal(t, "parent-123", link.ParentID)
	assert.Equal(t, "child-456", link.ChildID)
	assert.Equal(t, "board-123", link.BoardID)
	assert.Equal(t, domain.LinkTypeSubtask, link.LinkType)
	assert.NotEmpty(t, link.ID)

	cardLinkRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestLinkCards_SelfLink_Error(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	cardRepo := new(MockCardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)

	uc := NewLinkCardsUseCase(cardLinkRepo, cardRepo, memberRepo, publisher)
	link, err := uc.Execute(context.Background(), "card-123", "card-123", "board-123", "user-123")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrSelfLink, err)
	assert.Nil(t, link)

	memberRepo.AssertExpectations(t)
}

func TestLinkCards_NonMember_Denied(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	cardRepo := new(MockCardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewLinkCardsUseCase(cardLinkRepo, cardRepo, memberRepo, publisher)
	link, err := uc.Execute(context.Background(), "parent-123", "child-456", "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, link)

	memberRepo.AssertExpectations(t)
}

func TestLinkCards_AlreadyExists_Error(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	cardRepo := new(MockCardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cardLinkRepo.On("CreateVerified", mock.Anything, mock.AnythingOfType("*domain.CardLink")).
		Return(domain.ErrLinkAlreadyExists)

	uc := NewLinkCardsUseCase(cardLinkRepo, cardRepo, memberRepo, publisher)
	link, err := uc.Execute(context.Background(), "parent-123", "child-456", "board-123", "user-123")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrLinkAlreadyExists, err)
	assert.Nil(t, link)

	cardLinkRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestLinkCards_ParentNotFound_Error(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	cardRepo := new(MockCardRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cardLinkRepo.On("CreateVerified", mock.Anything, mock.AnythingOfType("*domain.CardLink")).
		Return(domain.ErrCardNotFound)

	uc := NewLinkCardsUseCase(cardLinkRepo, cardRepo, memberRepo, publisher)
	link, err := uc.Execute(context.Background(), "parent-999", "child-456", "board-123", "user-123")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCardNotFound, err)
	assert.Nil(t, link)

	cardLinkRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestUnlinkCards_Success(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cardLinkRepo.On("GetByID", mock.Anything, "link-123", "board-123").
		Return(&domain.CardLink{
			ID:       "link-123",
			ParentID: "parent-123",
			ChildID:  "child-456",
			BoardID:  "board-123",
			LinkType: domain.LinkTypeSubtask,
		}, nil)
	cardLinkRepo.On("Delete", mock.Anything, "link-123", "board-123").Return(nil)
	publisher.On("PublishCardUnlinked", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewUnlinkCardsUseCase(cardLinkRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "link-123", "board-123", "user-123")

	assert.NoError(t, err)

	cardLinkRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestUnlinkCards_NonMember_Denied(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewUnlinkCardsUseCase(cardLinkRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "link-123", "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)

	memberRepo.AssertExpectations(t)
}

func TestUnlinkCards_LinkNotFound_Error(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cardLinkRepo.On("GetByID", mock.Anything, "link-999", "board-123").
		Return(nil, domain.ErrCardLinkNotFound)

	uc := NewUnlinkCardsUseCase(cardLinkRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "link-999", "board-123", "user-123")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCardLinkNotFound, err)

	cardLinkRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetCardChildren_Success(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cardLinkRepo.On("ListChildren", mock.Anything, "parent-123", "board-123").
		Return([]*domain.CardLink{
			{ID: "link-1", ParentID: "parent-123", ChildID: "child-1", BoardID: "board-123", LinkType: domain.LinkTypeSubtask},
			{ID: "link-2", ParentID: "parent-123", ChildID: "child-2", BoardID: "board-123", LinkType: domain.LinkTypeSubtask},
		}, nil)

	uc := NewGetCardChildrenUseCase(cardLinkRepo, memberRepo)
	links, err := uc.Execute(context.Background(), "parent-123", "board-123", "user-123")

	assert.NoError(t, err)
	assert.Len(t, links, 2)
	assert.Equal(t, "child-1", links[0].ChildID)
	assert.Equal(t, "child-2", links[1].ChildID)

	cardLinkRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetCardChildren_NonMember_Denied(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewGetCardChildrenUseCase(cardLinkRepo, memberRepo)
	links, err := uc.Execute(context.Background(), "parent-123", "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, links)

	memberRepo.AssertExpectations(t)
}

func TestGetCardParents_Success(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cardLinkRepo.On("ListParents", mock.Anything, "child-456").
		Return([]*domain.CardLink{
			{ID: "link-1", ParentID: "parent-1", ChildID: "child-456", BoardID: "board-123", LinkType: domain.LinkTypeSubtask},
		}, nil)

	uc := NewGetCardParentsUseCase(cardLinkRepo, memberRepo)
	links, err := uc.Execute(context.Background(), "child-456", "board-123", "user-123")

	assert.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, "parent-1", links[0].ParentID)

	cardLinkRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetCardParents_NonMember_Denied(t *testing.T) {
	cardLinkRepo := new(MockCardLinkRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewGetCardParentsUseCase(cardLinkRepo, memberRepo)
	links, err := uc.Execute(context.Background(), "child-456", "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, links)

	memberRepo.AssertExpectations(t)
}
