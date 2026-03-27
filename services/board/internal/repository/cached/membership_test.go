package cached

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/cache"
)

// --- Mock PostgreSQL repo ---

type mockPgRepo struct {
	mock.Mock
}

func (m *mockPgRepo) IsMember(ctx context.Context, boardID, userID string) (bool, domain.Role, error) {
	args := m.Called(ctx, boardID, userID)
	return args.Bool(0), args.Get(1).(domain.Role), args.Error(2)
}

func (m *mockPgRepo) AddMember(ctx context.Context, boardID, userID string, role domain.Role) error {
	return m.Called(ctx, boardID, userID, role).Error(0)
}

func (m *mockPgRepo) RemoveMember(ctx context.Context, boardID, userID string) error {
	return m.Called(ctx, boardID, userID).Error(0)
}

func (m *mockPgRepo) ListMembers(ctx context.Context, boardID string, limit, offset int) ([]*domain.Member, error) {
	args := m.Called(ctx, boardID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Member), args.Error(1)
}

// --- Helpers ---

func setupCachedRepo(t *testing.T) (*MembershipRepository, *mockPgRepo, *cache.MembershipCache) {
	t.Helper()
	mr := miniredis.RunT(t)
	c, err := cache.NewMembershipCache("redis://" + mr.Addr())
	require.NoError(t, err)
	pg := new(mockPgRepo)
	repo := NewMembershipRepository(pg, c)
	return repo, pg, c
}

// --- Tests ---

func TestIsMember_CacheHit(t *testing.T) {
	repo, pg, c := setupCachedRepo(t)
	ctx := context.Background()

	// Предзаполняем кеш (как это сделал бы cache consumer)
	_ = c.SetMember(ctx, "board-1", "user-1", "owner")

	isMember, role, err := repo.IsMember(ctx, "board-1", "user-1")
	require.NoError(t, err)
	assert.True(t, isMember)
	assert.Equal(t, domain.RoleOwner, role)

	// PostgreSQL НЕ вызван
	pg.AssertNotCalled(t, "IsMember")
}

func TestIsMember_CacheMiss_FallbackToPostgres(t *testing.T) {
	repo, pg, _ := setupCachedRepo(t)
	ctx := context.Background()

	// Кеш пустой — PostgreSQL вернёт результат
	pg.On("IsMember", ctx, "board-1", "user-1").Return(true, domain.RoleMember, nil)

	isMember, role, err := repo.IsMember(ctx, "board-1", "user-1")
	require.NoError(t, err)
	assert.True(t, isMember)
	assert.Equal(t, domain.RoleMember, role)

	pg.AssertCalled(t, "IsMember", ctx, "board-1", "user-1")
}

func TestIsMember_CacheMiss_NotMember(t *testing.T) {
	repo, pg, _ := setupCachedRepo(t)
	ctx := context.Background()

	pg.On("IsMember", ctx, "board-1", "outsider").Return(false, domain.Role(""), nil)

	isMember, _, err := repo.IsMember(ctx, "board-1", "outsider")
	require.NoError(t, err)
	assert.False(t, isMember)

	pg.AssertCalled(t, "IsMember", ctx, "board-1", "outsider")
}

func TestIsMember_CacheError_FallbackToPostgres(t *testing.T) {
	repo, pg, c := setupCachedRepo(t)
	ctx := context.Background()

	// Закрываем Redis чтобы вызвать ошибку
	_ = c.Close()

	pg.On("IsMember", ctx, "board-1", "user-1").Return(true, domain.RoleOwner, nil)

	isMember, role, err := repo.IsMember(ctx, "board-1", "user-1")
	require.NoError(t, err)
	assert.True(t, isMember)
	assert.Equal(t, domain.RoleOwner, role)

	// PostgreSQL вызван как fallback
	pg.AssertCalled(t, "IsMember", ctx, "board-1", "user-1")
}

func TestAddMember_DelegatesToPostgres(t *testing.T) {
	repo, pg, _ := setupCachedRepo(t)
	ctx := context.Background()

	pg.On("AddMember", ctx, "board-1", "user-2", domain.RoleMember).Return(nil)

	err := repo.AddMember(ctx, "board-1", "user-2", domain.RoleMember)
	require.NoError(t, err)

	pg.AssertCalled(t, "AddMember", ctx, "board-1", "user-2", domain.RoleMember)
}

func TestRemoveMember_DelegatesToPostgres(t *testing.T) {
	repo, pg, _ := setupCachedRepo(t)
	ctx := context.Background()

	pg.On("RemoveMember", ctx, "board-1", "user-2").Return(nil)

	err := repo.RemoveMember(ctx, "board-1", "user-2")
	require.NoError(t, err)

	pg.AssertCalled(t, "RemoveMember", ctx, "board-1", "user-2")
}

func TestListMembers_DelegatesToPostgres(t *testing.T) {
	repo, pg, _ := setupCachedRepo(t)
	ctx := context.Background()

	members := []*domain.Member{{UserID: "user-1"}}
	pg.On("ListMembers", ctx, "board-1", 10, 0).Return(members, nil)

	result, err := repo.ListMembers(ctx, "board-1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, result, 1)

	pg.AssertCalled(t, "ListMembers", ctx, "board-1", 10, 0)
}

func TestIsMember_PostgresError_Propagated(t *testing.T) {
	repo, pg, _ := setupCachedRepo(t)
	ctx := context.Background()

	pg.On("IsMember", ctx, "board-1", "user-1").Return(false, domain.Role(""), errors.New("db error"))

	_, _, err := repo.IsMember(ctx, "board-1", "user-1")
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}
