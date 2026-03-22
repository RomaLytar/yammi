package cache

import (
	"context"
	"sync"
)

// NameCacheRepo — интерфейс PostgreSQL name cache repository.
type NameCacheRepo interface {
	SetBoardName(ctx context.Context, boardID, title string) error
	GetBoardName(ctx context.Context, boardID string) string
	DeleteBoardName(ctx context.Context, boardID string) error
	SetUserName(ctx context.Context, userID, name string) error
	GetUserName(ctx context.Context, userID string) string
	SetCardName(ctx context.Context, cardID, title string) error
	GetCardName(ctx context.Context, cardID string) string
	DeleteCardName(ctx context.Context, cardID string) error
	SetColumnName(ctx context.Context, columnID, title string) error
	GetColumnName(ctx context.Context, columnID string) string
	DeleteColumnName(ctx context.Context, columnID string) error
	TruncateCache(ctx context.Context) error
}

// InMemoryNameCache — in-memory cache с write-through в PostgreSQL.
// Get: O(1) из памяти (0 DB queries на write path).
// Set: memory + async DB write.
type InMemoryNameCache struct {
	delegate NameCacheRepo
	mu       sync.RWMutex
	boards   map[string]string
	users    map[string]string
	cards    map[string]string
	columns  map[string]string
}

func NewInMemoryNameCache(delegate NameCacheRepo) *InMemoryNameCache {
	return &InMemoryNameCache{
		delegate: delegate,
		boards:   make(map[string]string),
		users:    make(map[string]string),
		cards:    make(map[string]string),
		columns:  make(map[string]string),
	}
}

// --- Board names ---

func (c *InMemoryNameCache) SetBoardName(ctx context.Context, boardID, title string) error {
	c.mu.Lock()
	c.boards[boardID] = title
	c.mu.Unlock()
	go c.delegate.SetBoardName(context.Background(), boardID, title)
	return nil
}

func (c *InMemoryNameCache) GetBoardName(ctx context.Context, boardID string) string {
	c.mu.RLock()
	title, ok := c.boards[boardID]
	c.mu.RUnlock()
	if ok {
		return title
	}
	// Fallback to DB
	title = c.delegate.GetBoardName(ctx, boardID)
	if title != "" {
		c.mu.Lock()
		c.boards[boardID] = title
		c.mu.Unlock()
	}
	return title
}

func (c *InMemoryNameCache) DeleteBoardName(ctx context.Context, boardID string) error {
	c.mu.Lock()
	delete(c.boards, boardID)
	c.mu.Unlock()
	go c.delegate.DeleteBoardName(context.Background(), boardID)
	return nil
}

// --- User names ---

func (c *InMemoryNameCache) SetUserName(ctx context.Context, userID, name string) error {
	c.mu.Lock()
	c.users[userID] = name
	c.mu.Unlock()
	go c.delegate.SetUserName(context.Background(), userID, name)
	return nil
}

func (c *InMemoryNameCache) GetUserName(ctx context.Context, userID string) string {
	c.mu.RLock()
	name, ok := c.users[userID]
	c.mu.RUnlock()
	if ok {
		return name
	}
	name = c.delegate.GetUserName(ctx, userID)
	if name != "" {
		c.mu.Lock()
		c.users[userID] = name
		c.mu.Unlock()
	}
	return name
}

// --- Card names ---

func (c *InMemoryNameCache) SetCardName(ctx context.Context, cardID, title string) error {
	c.mu.Lock()
	c.cards[cardID] = title
	c.mu.Unlock()
	go c.delegate.SetCardName(context.Background(), cardID, title)
	return nil
}

func (c *InMemoryNameCache) GetCardName(ctx context.Context, cardID string) string {
	c.mu.RLock()
	title, ok := c.cards[cardID]
	c.mu.RUnlock()
	if ok {
		return title
	}
	title = c.delegate.GetCardName(ctx, cardID)
	if title != "" {
		c.mu.Lock()
		c.cards[cardID] = title
		c.mu.Unlock()
	}
	return title
}

func (c *InMemoryNameCache) DeleteCardName(ctx context.Context, cardID string) error {
	c.mu.Lock()
	delete(c.cards, cardID)
	c.mu.Unlock()
	go c.delegate.DeleteCardName(context.Background(), cardID)
	return nil
}

// --- Column names ---

func (c *InMemoryNameCache) SetColumnName(ctx context.Context, columnID, title string) error {
	c.mu.Lock()
	c.columns[columnID] = title
	c.mu.Unlock()
	go c.delegate.SetColumnName(context.Background(), columnID, title)
	return nil
}

func (c *InMemoryNameCache) GetColumnName(ctx context.Context, columnID string) string {
	c.mu.RLock()
	title, ok := c.columns[columnID]
	c.mu.RUnlock()
	if ok {
		return title
	}
	title = c.delegate.GetColumnName(ctx, columnID)
	if title != "" {
		c.mu.Lock()
		c.columns[columnID] = title
		c.mu.Unlock()
	}
	return title
}

func (c *InMemoryNameCache) DeleteColumnName(ctx context.Context, columnID string) error {
	c.mu.Lock()
	delete(c.columns, columnID)
	c.mu.Unlock()
	go c.delegate.DeleteColumnName(context.Background(), columnID)
	return nil
}

func (c *InMemoryNameCache) TruncateCache(ctx context.Context) error {
	c.mu.Lock()
	c.boards = make(map[string]string)
	c.users = make(map[string]string)
	c.cards = make(map[string]string)
	c.columns = make(map[string]string)
	c.mu.Unlock()
	return c.delegate.TruncateCache(ctx)
}
