package snapshot

import (
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"
)

func BenchmarkEvaluateParallel(b *testing.B) {
	store := NewStore(&Snapshot{Revision: 1, Flags: map[string]Flag{
		"checkout": {Key: "checkout", Enabled: true},
	}})
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			store.Evaluate("checkout")
		}
	})
}

func TestEvaluateMillionConcurrentOperations(t *testing.T) {
	const operations = 1_000_000
	store := NewStore(&Snapshot{Revision: 1, Flags: map[string]Flag{
		"checkout": {Key: "checkout", Enabled: true},
	}})
	workers := runtime.GOMAXPROCS(0) * 4
	if workers > operations {
		workers = operations
	}
	latencies := make([]int64, operations)
	var group sync.WaitGroup
	group.Add(workers)
	start := time.Now()
	for worker := 0; worker < workers; worker++ {
		first := worker * operations / workers
		last := (worker + 1) * operations / workers
		go func(first, last int) {
			defer group.Done()
			for i := first; i < last; i++ {
				begin := time.Now()
				enabled, found, _ := store.Evaluate("checkout")
				latencies[i] = time.Since(begin).Nanoseconds()
				if !enabled || !found {
					t.Errorf("unexpected result enabled=%v found=%v", enabled, found)
					return
				}
			}
		}(first, last)
	}
	group.Wait()
	elapsed := time.Since(start)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	t.Logf("operations=%d workers=%d total=%s throughput=%.0f ops/s p50=%s p95=%s p99=%s max=%s",
		operations, workers, elapsed, float64(operations)/elapsed.Seconds(),
		time.Duration(percentile(latencies, 0.50)),
		time.Duration(percentile(latencies, 0.95)),
		time.Duration(percentile(latencies, 0.99)),
		time.Duration(latencies[len(latencies)-1]),
	)
}

func percentile(sorted []int64, p float64) int64 {
	return sorted[int(float64(len(sorted)-1)*p)]
}
