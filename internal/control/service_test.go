package control

import (
	"fmt"
	"sync"
	"testing"

	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

func TestServiceCRUDPublishesToDataPlane(t *testing.T) {
	data := snapshot.NewStore(nil)
	service := NewService(data)
	created, err := service.Create("checkout", true)
	if err != nil || created.Revision != 1 {
		t.Fatalf("create got flag=%+v err=%v", created, err)
	}
	updated, err := service.Update("checkout", false)
	if err != nil || updated.Revision != 2 || updated.Enabled {
		t.Fatalf("update got flag=%+v err=%v", updated, err)
	}
	enabled, found, revision := data.Evaluate("checkout")
	if enabled || !found || revision != 2 {
		t.Fatalf("data plane got enabled=%v found=%v revision=%d", enabled, found, revision)
	}
	if err := service.Delete("checkout"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get("checkout"); err != ErrNotFound {
		t.Fatalf("get after delete err=%v", err)
	}
}

func TestServiceConcurrentCreatesUseUniqueRevisions(t *testing.T) {
	service := NewService(nil)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = service.Create(fmt.Sprintf("flag-%03d", i), true)
		}(i)
	}
	wg.Wait()
	if got := len(service.List()); got != 100 {
		t.Fatalf("got %d flags, want 100", got)
	}
}
