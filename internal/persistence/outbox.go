package persistence

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type OutboxEvent struct {
	ID        int64
	Revision  uint64
	EventType string
	Payload   []byte
	Attempts  int
}

func (s *Store) ClaimPending(ctx context.Context, limit int) ([]OutboxEvent, error) {
	if limit < 1 {
		limit = 1
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id, revision, event_type, payload, attempts
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY id
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var event OutboxEvent
		if err := rows.Scan(&event.ID, &event.Revision, &event.EventType, &event.Payload, &event.Attempts); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, event := range events {
		if _, err := tx.Exec(ctx, `UPDATE outbox_events SET attempts = attempts + 1 WHERE id = $1`, event.ID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Store) MarkPublished(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE outbox_events SET published_at = now() WHERE id = $1 AND published_at IS NULL`, id)
	return err
}

type Publisher interface {
	Publish(context.Context, OutboxEvent) error
}

type Worker struct {
	store     *Store
	publisher Publisher
	batch     int
	interval  time.Duration
}

func NewWorker(store *Store, publisher Publisher, batch int, interval time.Duration) *Worker {
	if batch < 1 {
		batch = 100
	}
	if interval <= 0 {
		interval = time.Second
	}
	return &Worker{store: store, publisher: publisher, batch: batch, interval: interval}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	events, err := w.store.ClaimPending(ctx, w.batch)
	if err != nil {
		return err
	}
	for _, event := range events {
		if err := w.publisher.Publish(ctx, event); err != nil {
			continue
		}
		if err := w.store.MarkPublished(ctx, event.ID); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		if err := w.RunOnce(ctx); err != nil && ctx.Err() != nil {
			return ctx.Err()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
