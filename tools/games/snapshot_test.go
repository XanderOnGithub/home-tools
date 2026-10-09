package games

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSnapshot(t *testing.T) {
	var s snapshot
	var calls atomic.Int32
	gate := make(chan struct{}) // lets the test decide when a lookup finishes
	build := func() []serverView {
		n := calls.Add(1)
		<-gate
		return []serverView{{ID: "mc", Name: string('A' + rune(n) - 1)}}
	}
	ctx := context.Background()

	// First requests wait, and share one lookup.
	var wg sync.WaitGroup
	results := make([][]serverView, 2)
	for i := range results {
		wg.Go(func() { results[i] = s.get(ctx, build) })
	}
	time.Sleep(20 * time.Millisecond)
	gate <- struct{}{}
	wg.Wait()
	if calls.Load() != 1 || results[0][0].Name != "A" || results[1][0].Name != "A" {
		t.Fatalf("first: calls=%d results=%+v", calls.Load(), results)
	}

	// Fresh: served from memory.
	if got := s.get(ctx, build); calls.Load() != 1 || got[0].Name != "A" {
		t.Fatalf("fresh: calls=%d got=%+v", calls.Load(), got)
	}

	// Stale: the old views come back at once; one refresh starts.
	s.mu.Lock()
	s.takenAt = time.Now().Add(-time.Minute)
	s.mu.Unlock()
	start := time.Now()
	if got := s.get(ctx, build); got[0].Name != "A" || time.Since(start) > 50*time.Millisecond {
		t.Fatalf("stale: got=%+v after %v", got, time.Since(start))
	}
	s.get(ctx, build) // still refreshing: no second refresh
	gate <- struct{}{}
	time.Sleep(20 * time.Millisecond)
	if got := s.get(ctx, build); calls.Load() != 2 || got[0].Name != "B" {
		t.Fatalf("after refresh: calls=%d got=%+v", calls.Load(), got)
	}

	// put replaces one view in place.
	s.put(serverView{ID: "mc", Name: "after action"})
	if got := s.get(ctx, build); got[0].Name != "after action" {
		t.Fatalf("put: got=%+v", got)
	}

	// reset: the next request waits for a full lookup again.
	s.reset()
	done := make(chan []serverView)
	go func() { done <- s.get(ctx, build) }()
	select {
	case <-done:
		t.Fatal("after reset, get returned before the lookup finished")
	case <-time.After(20 * time.Millisecond):
	}
	gate <- struct{}{}
	if got := <-done; got[0].Name != "C" {
		t.Fatalf("after reset: got=%+v", got)
	}
}
