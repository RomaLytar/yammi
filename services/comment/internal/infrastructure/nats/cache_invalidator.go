package nats

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

// memberRemovedEvent — минимальная структура события member.removed.
type memberRemovedEvent struct {
	BoardID string `json:"board_id"`
	UserID  string `json:"user_id"`
}

// CacheInvalidator — интерфейс для инвалидации кэша членства.
type CacheInvalidator interface {
	InvalidateCache(boardID, userID string)
}

// SubscribeMemberRemoved подписывается на member.removed и инвалидирует кэш членства,
// чтобы удалённый участник не мог пройти проверку по stale-кэшу.
func SubscribeMemberRemoved(conn *nats.Conn, cache CacheInvalidator) (*nats.Subscription, error) {
	sub, err := conn.Subscribe("member.removed", func(msg *nats.Msg) {
		var evt memberRemovedEvent
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Printf("cache-invalidator: failed to parse member.removed: %v", err)
			return
		}

		if evt.BoardID != "" && evt.UserID != "" {
			cache.InvalidateCache(evt.BoardID, evt.UserID)
			log.Printf("cache-invalidator: invalidated membership cache for user=%s board=%s", evt.UserID, evt.BoardID)
		}
	})
	if err != nil {
		return nil, err
	}

	log.Println("cache-invalidator: subscribed to member.removed")
	return sub, nil
}
