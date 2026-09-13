package persistence

import "testing"

func TestWorkerDefaultsAreSafe(t *testing.T) {
	worker := NewWorker(nil, nil, 0, 0)
	if worker.batch != 100 {
		t.Fatalf("batch=%d, want 100", worker.batch)
	}
	if worker.interval <= 0 {
		t.Fatal("worker interval must be positive")
	}
}
