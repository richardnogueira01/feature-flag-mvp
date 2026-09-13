package messaging

import (
	"context"
	"encoding/json"
	"strings"

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
	if _, err := s.stream.AddStream(&nats.StreamConfig{
		Name:     streamName(subject),
		Subjects: []string{subject},
		Storage:  nats.FileStorage,
	}); err != nil && !strings.Contains(err.Error(), "stream name already in use") {
		return nil, err
	}
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

func streamName(subject string) string {
	name := strings.NewReplacer(".", "_", "*", "ALL", ">", "REST").Replace(subject)
	return "FEATURE_FLAGS_" + strings.ToUpper(name)
}
