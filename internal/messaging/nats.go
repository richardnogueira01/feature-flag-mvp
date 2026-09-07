package messaging

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/richardnogueira01/feature-flag-mvp/internal/persistence"
)

type Publisher struct {
	stream  nats.JetStreamContext
	subject string
}

func NewPublisher(stream nats.JetStreamContext, subject string) *Publisher {
	return &Publisher{stream: stream, subject: subject}
}

func (p *Publisher) Publish(ctx context.Context, event persistence.OutboxEvent) error {
	payload, err := json.Marshal(struct {
		ID        int64  `json:"id"`
		Revision  uint64 `json:"revision"`
		EventType string `json:"event_type"`
		Payload   []byte `json:"payload"`
	}{event.ID, event.Revision, event.EventType, event.Payload})
	if err != nil {
		return err
	}
	_, err = p.stream.Publish(p.subject, payload, nats.Context(ctx))
	return err
}
