package snapshot

import (
	"sync"
	"testing"
)

func TestStoreEvaluate(t *testing.T) {
	store := NewStore(&Snapshot{
		Revision: 1,
		Flags:    map[string]Flag{"checkout": {Key: "checkout", Enabled: true}},
	})

	enabled, found, revision := store.Evaluate("checkout")
	if !enabled || !found || revision != 1 {
		t.Fatalf("got enabled=%v found=%v revision=%d", enabled, found, revision)
	}

	enabled, found, revision = store.Evaluate("missing")
	if enabled || found || revision != 1 {
		t.Fatalf("missing flag got enabled=%v found=%v revision=%d", enabled, found, revision)
	}
}

func TestStoreDoesNotRetainMutableInputMap(t *testing.T) {
	flags := map[string]Flag{"billing": {Key: "billing", Enabled: true}}
	store := NewStore(&Snapshot{Revision: 1, Flags: flags})
	flags["billing"] = Flag{Key: "billing", Enabled: false}

	enabled, _, _ := store.Evaluate("billing")
	if !enabled {
		t.Fatal("published snapshot changed after input map mutation")
	}
}

func TestStoreSupportsConcurrentEvaluationAndPublication(t *testing.T) {
	store := NewStore(&Snapshot{Revision: 1, Flags: map[string]Flag{"search": {Key: "search"}}})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				store.Evaluate("search")
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for revision := uint64(2); revision <= 100; revision++ {
			store.Publish(&Snapshot{Revision: revision, Flags: map[string]Flag{
				"search": {Key: "search", Enabled: revision%2 == 0},
			}})
		}
	}()
	wg.Wait()
}
