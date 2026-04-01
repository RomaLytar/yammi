package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
)

// memberCacheEntry — запись в кеше членства
type memberCacheEntry struct {
	isMember  bool
	isOwner   bool
	expiresAt time.Time
}

// BoardMembershipChecker проверяет членство через Board Service gRPC
type BoardMembershipChecker struct {
	client boardpb.BoardServiceClient
	conn   *grpc.ClientConn

	mu    sync.RWMutex
	cache map[string]memberCacheEntry // key: "boardID:userID"
	ttl   time.Duration
}

// grpcSecretClientInterceptor добавляет shared secret в metadata исходящих gRPC вызовов
// для аутентификации между внутренними сервисами.
func grpcSecretClientInterceptor(secret string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if secret != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-internal-secret", secret)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// NewBoardMembershipChecker создает клиент для проверки членства
func NewBoardMembershipChecker(boardGRPCAddr, sharedSecret string) (*BoardMembershipChecker, error) {
	conn, err := grpc.NewClient(boardGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcSecretClientInterceptor(sharedSecret)),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to board service: %w", err)
	}

	return &BoardMembershipChecker{
		client: boardpb.NewBoardServiceClient(conn),
		conn:   conn,
		cache:  make(map[string]memberCacheEntry),
		ttl:    30 * time.Second,
	}, nil
}

// IsMember проверяет, является ли пользователь членом доски.
// Кэшируются только положительные результаты (isMember=true) — отказ в доступе
// всегда перепроверяется через Board Service, чтобы отзыв прав действовал немедленно.
func (c *BoardMembershipChecker) IsMember(ctx context.Context, boardID, userID string) (bool, error) {
	entry, ok := c.getCached(boardID, userID)
	if ok {
		return entry.isMember, nil
	}

	isMember, isOwner, err := c.fetchMembership(ctx, boardID, userID)
	if err != nil {
		return false, err
	}

	if isMember {
		c.setCached(boardID, userID, isMember, isOwner)
	}
	return isMember, nil
}

// IsOwner проверяет, является ли пользователь владельцем доски.
func (c *BoardMembershipChecker) IsOwner(ctx context.Context, boardID, userID string) (bool, error) {
	entry, ok := c.getCached(boardID, userID)
	if ok {
		return entry.isOwner, nil
	}

	isMember, isOwner, err := c.fetchMembership(ctx, boardID, userID)
	if err != nil {
		return false, err
	}

	if isMember {
		c.setCached(boardID, userID, isMember, isOwner)
	}
	return isOwner, nil
}

// fetchMembership запрашивает Board Service IsMember RPC
func (c *BoardMembershipChecker) fetchMembership(ctx context.Context, boardID, userID string) (isMember bool, isOwner bool, err error) {
	resp, err := c.client.IsMember(ctx, &boardpb.IsMemberRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		return false, false, fmt.Errorf("check membership via board service: %w", err)
	}

	return resp.GetIsMember(), resp.GetRole() == "owner", nil
}

func (c *BoardMembershipChecker) getCached(boardID, userID string) (memberCacheEntry, bool) {
	key := boardID + ":" + userID
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.cache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return memberCacheEntry{}, false
	}
	return entry, true
}

func (c *BoardMembershipChecker) setCached(boardID, userID string, isMember, isOwner bool) {
	key := boardID + ":" + userID
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = memberCacheEntry{
		isMember:  isMember,
		isOwner:   isOwner,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// CardExistsInBoard проверяет, что карточка существует и принадлежит указанной доске.
// Вызывает Board Service GetCard RPC.
// NotFound/PermissionDenied → (false, nil); инфраструктурные ошибки → (false, err).
func (c *BoardMembershipChecker) CardExistsInBoard(ctx context.Context, cardID, boardID, userID string) (bool, error) {
	_, err := c.client.GetCard(ctx, &boardpb.GetCardRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		st, ok := grpcstatus.FromError(err)
		if ok {
			switch st.Code() {
			case grpccodes.NotFound, grpccodes.PermissionDenied:
				return false, nil
			}
		}
		return false, fmt.Errorf("check card existence via board service: %w", err)
	}
	return true, nil
}

// InvalidateCache удаляет кэшированный результат для пары boardID:userID.
func (c *BoardMembershipChecker) InvalidateCache(boardID, userID string) {
	key := boardID + ":" + userID
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

// Close закрывает gRPC соединение
func (c *BoardMembershipChecker) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
