package cache

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestCache(t *testing.T) (*MembershipCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	cache, err := NewMembershipCache("redis://" + mr.Addr())
	require.NoError(t, err)
	return cache, mr
}

func TestSetMember_GetRole(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	err := cache.SetMember(ctx, "board-1", "user-1", "owner")
	require.NoError(t, err)

	role, found, err := cache.GetRole(ctx, "board-1", "user-1")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "owner", role)
}

func TestSetMember_MultipleMembersOneBoard(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	_ = cache.SetMember(ctx, "board-1", "user-1", "owner")
	_ = cache.SetMember(ctx, "board-1", "user-2", "member")
	_ = cache.SetMember(ctx, "board-1", "user-3", "member")

	role1, found1, _ := cache.GetRole(ctx, "board-1", "user-1")
	assert.True(t, found1)
	assert.Equal(t, "owner", role1)

	role2, found2, _ := cache.GetRole(ctx, "board-1", "user-2")
	assert.True(t, found2)
	assert.Equal(t, "member", role2)

	role3, found3, _ := cache.GetRole(ctx, "board-1", "user-3")
	assert.True(t, found3)
	assert.Equal(t, "member", role3)
}

func TestGetRole_NotFound(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	_, found, err := cache.GetRole(ctx, "board-1", "user-1")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestRemoveMember(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	_ = cache.SetMember(ctx, "board-1", "user-1", "owner")
	_ = cache.SetMember(ctx, "board-1", "user-2", "member")

	err := cache.RemoveMember(ctx, "board-1", "user-2")
	require.NoError(t, err)

	// user-2 удалён
	_, found, _ := cache.GetRole(ctx, "board-1", "user-2")
	assert.False(t, found)

	// user-1 на месте
	role, found, _ := cache.GetRole(ctx, "board-1", "user-1")
	assert.True(t, found)
	assert.Equal(t, "owner", role)
}

func TestRemoveBoard_CleansAllData(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	_ = cache.SetMember(ctx, "board-1", "user-1", "owner")
	_ = cache.SetMember(ctx, "board-1", "user-2", "member")
	_ = cache.SetMember(ctx, "board-1", "user-3", "member")

	// Проверяем board_roles до удаления
	role1, _, _ := cache.GetRole(ctx, "board-1", "user-1")
	assert.Equal(t, "owner", role1)

	err := cache.RemoveBoard(ctx, "board-1")
	require.NoError(t, err)

	// board_roles удалён
	_, found, _ := cache.GetRole(ctx, "board-1", "user-1")
	assert.False(t, found)

	// user_boards очищены
	_, found2, _ := cache.GetRole(ctx, "board-1", "user-2")
	assert.False(t, found2)
	_, found3, _ := cache.GetRole(ctx, "board-1", "user-3")
	assert.False(t, found3)
}

func TestSetMember_UpdatesUserBoards(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	_ = cache.SetMember(ctx, "board-1", "user-1", "owner")
	_ = cache.SetMember(ctx, "board-2", "user-1", "member")
	_ = cache.SetMember(ctx, "board-3", "user-1", "member")

	// user-1 должен быть в 3 досках — проверяем через GetRole
	_, f1, _ := cache.GetRole(ctx, "board-1", "user-1")
	_, f2, _ := cache.GetRole(ctx, "board-2", "user-1")
	_, f3, _ := cache.GetRole(ctx, "board-3", "user-1")
	assert.True(t, f1)
	assert.True(t, f2)
	assert.True(t, f3)
}

func TestRemoveMember_UpdatesUserBoards(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	_ = cache.SetMember(ctx, "board-1", "user-1", "owner")
	_ = cache.SetMember(ctx, "board-2", "user-1", "member")
	_ = cache.RemoveMember(ctx, "board-2", "user-1")

	_, fb1, _ := cache.GetRole(ctx, "board-1", "user-1")
	_, fb2, _ := cache.GetRole(ctx, "board-2", "user-1")
	assert.True(t, fb1)
	assert.False(t, fb2)
}

func TestFlush(t *testing.T) {
	cache, _ := setupTestCache(t)
	ctx := context.Background()

	_ = cache.SetMember(ctx, "board-1", "user-1", "owner")

	err := cache.Flush(ctx)
	require.NoError(t, err)

	_, found, _ := cache.GetRole(ctx, "board-1", "user-1")
	assert.False(t, found)
}
