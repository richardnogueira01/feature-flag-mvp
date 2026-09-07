package messaging

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

type eventEnvelope struct {
	Revision uint64 `json:"revision"`
	Payload  []byte `json:"payload"`
}

type Subscriber struct {
	stream nats.JetStreamContext
}

func NewSubscriber(stream nats.JetStreamContext) *Subscriber {
	return &Subscriber{stream: stream}
}

func (s *Subscriber) Subscribe(subject, durable string, apply func(context.Context, *snapshot.Snapshot) error) (*nats.Subscription, error) {
	return s.stream.Subscribe(subject, func(message *nats.Msg) {
		var envelope eventEnvelope
		if err := json.Unmarshal(message.Data, &envelope); err != nil {
			_ = message.Term()
			return
		}
		var next snapshot.Snapshot
		if err := json.Unmarshal(envelope.Payload, &next); err != nil {
			_ = message.Term()
			return
		}
		next.Revision = envelope.Revision
		if err := apply(context.Background(), &next); err != nil {
			_ = message.Nak()
			return
		}
		_ = message.Ack()
	}, nats.Durable(durable), nats.ManualAck())
}
