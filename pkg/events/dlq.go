package events

import "time"

const (
	StreamDLQ        = "DLQ"
	SubjectDLQPrefix = "dlq."
)

func DLQSubject(originalSubject string) string {
	return SubjectDLQPrefix + originalSubject
}

type DLQEnvelope struct {
	OriginalSubject string `json:"original_subject"`
	ConsumerName    string `json:"consumer_name"`
	Error           string `json:"error"`
	NumDelivered    uint64 `json:"num_delivered"`
	Payload         string `json:"payload"`
	FailedAt        time.Time `json:"failed_at"`
}
