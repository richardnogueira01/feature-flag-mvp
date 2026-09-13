package syncer

import (
	"context"
	"testing"

	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

func TestSyncerAppliesSequentialAndIgnoresOldRevisions(t *testing.T) {
	store := snapshot.NewStore(nil)
	syncer := New(store, nil)
	if err := syncer.Apply(context.Background(), &snapshot.Snapshot{Revision: 1, Flags: map[string]snapshot.Flag{"checkout": {Enabled: false}}}); err != nil {
		t.Fatal(err)
	}
	if err := syncer.Apply(context.Background(), &snapshot.Snapshot{Revision: 1, Flags: map[string]snapshot.Flag{"checkout": {Enabled: true}}}); err != nil {
		t.Fatal(err)
	}
	enabled, _, revision := store.Evaluate("checkout")
	if enabled || revision != 1 || !syncer.Ready() {
		t.Fatalf("enabled=%v revision=%d ready=%v", enabled, revision, syncer.Ready())
	}
}

func TestSyncerResyncsOnGap(t *testing.T) {
	store := snapshot.NewStore(&snapshot.Snapshot{Revision: 1})
	syncer := New(store, func(context.Context) (*snapshot.Snapshot, error) {
		return &snapshot.Snapshot{Revision: 4, Flags: map[string]snapshot.Flag{"search": {Enabled: true}}}, nil
	})
	if err := syncer.Apply(context.Background(), &snapshot.Snapshot{Revision: 3}); err != nil {
		t.Fatal(err)
	}
	if store.Revision() != 4 || !syncer.Ready() {
		t.Fatalf("revision=%d ready=%v", store.Revision(), syncer.Ready())
	}
}
