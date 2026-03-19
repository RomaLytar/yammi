// CLI утилита для работы с Dead Letter Queue.
//
// Команды:
//   list   — показать все сообщения в DLQ
//   replay — переотправить сообщения обратно в оригинальный стрим
//   purge  — очистить DLQ
//
// Использование:
//   go run tools/dlq/main.go list
//   go run tools/dlq/main.go replay
//   go run tools/dlq/main.go purge
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

// DLQEnvelope — копия pkg/events.DLQEnvelope (tools/dlq — отдельный Go module).
type DLQEnvelope struct {
	OriginalSubject string    `json:"original_subject"`
	ConsumerName    string    `json:"consumer_name"`
	Error           string    `json:"error"`
	NumDelivered    uint64    `json:"num_delivered"`
	Payload         string    `json:"payload"`
	FailedAt        time.Time `json:"failed_at"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dlq <list|replay|purge>")
		os.Exit(1)
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer nc.Close()

	js, _ := nc.JetStream()

	switch os.Args[1] {
	case "list":
		listDLQ(js)
	case "replay":
		replayDLQ(js)
	case "purge":
		purgeDLQ(js)
	default:
		fmt.Printf("Unknown command: %s\nUsage: dlq <list|replay|purge>\n", os.Args[1])
		os.Exit(1)
	}
}

func listDLQ(js nats.JetStreamContext) {
	info, err := js.StreamInfo("DLQ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "stream info: %v\n", err)
		return
	}

	total := info.State.Msgs
	if total == 0 {
		fmt.Println("DLQ is empty ✓")
		return
	}

	fmt.Printf("DLQ: %d messages\n\n", total)

	sub, err := js.SubscribeSync("dlq.>", nats.OrderedConsumer())
	if err != nil {
		fmt.Fprintf(os.Stderr, "subscribe: %v\n", err)
		return
	}
	defer sub.Unsubscribe()

	for i := uint64(0); i < total; i++ {
		msg, err := sub.NextMsg(5 * time.Second)
		if err != nil {
			break
		}

		var env DLQEnvelope
		if err := json.Unmarshal(msg.Data, &env); err != nil {
			fmt.Printf("#%d [RAW] subject=%s data=%s\n", i+1, msg.Subject, string(msg.Data))
			continue
		}

		fmt.Printf("#%d ──────────────────────────────────\n", i+1)
		fmt.Printf("  Subject:    %s\n", env.OriginalSubject)
		fmt.Printf("  Consumer:   %s\n", env.ConsumerName)
		fmt.Printf("  Error:      %s\n", env.Error)
		fmt.Printf("  Deliveries: %d\n", env.NumDelivered)
		fmt.Printf("  Failed at:  %s\n", env.FailedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Payload:    %s\n", env.Payload)
		fmt.Println()
	}
}

func replayDLQ(js nats.JetStreamContext) {
	info, err := js.StreamInfo("DLQ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "stream info: %v\n", err)
		return
	}

	total := info.State.Msgs
	if total == 0 {
		fmt.Println("DLQ is empty, nothing to replay ✓")
		return
	}

	fmt.Printf("Replaying %d messages from DLQ back to original subjects...\n\n", total)

	sub, err := js.SubscribeSync("dlq.>", nats.OrderedConsumer())
	if err != nil {
		fmt.Fprintf(os.Stderr, "subscribe: %v\n", err)
		return
	}
	defer sub.Unsubscribe()

	replayed := 0
	failed := 0

	for i := uint64(0); i < total; i++ {
		msg, err := sub.NextMsg(5 * time.Second)
		if err != nil {
			break
		}

		var env DLQEnvelope
		if err := json.Unmarshal(msg.Data, &env); err != nil {
			fmt.Printf("  SKIP #%d: can't parse envelope: %v\n", i+1, err)
			failed++
			continue
		}

		// Переотправляем оригинальный payload в оригинальный subject
		_, err = js.Publish(env.OriginalSubject, []byte(env.Payload))
		if err != nil {
			fmt.Printf("  FAIL #%d: publish to %s: %v\n", i+1, env.OriginalSubject, err)
			failed++
			continue
		}

		fmt.Printf("  ✓ #%d → %s (was: %s)\n", i+1, env.OriginalSubject, env.Error)
		replayed++
	}

	fmt.Printf("\nDone: %d replayed, %d failed\n", replayed, failed)

	if failed == 0 && replayed > 0 {
		fmt.Println("Run 'dlq purge' to clear processed DLQ messages")
	}
}

func purgeDLQ(js nats.JetStreamContext) {
	err := js.PurgeStream("DLQ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "purge: %v\n", err)
		return
	}
	fmt.Println("DLQ purged ✓")
}
