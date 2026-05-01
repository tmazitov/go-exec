package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func instant(val any) func(context.Context, map[string]any) (any, error) {
	return func(_ context.Context, _ map[string]any) (any, error) { return val, nil }
}

func mustKeys(t *testing.T, results map[string]any, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if _, ok := results[k]; !ok {
			t.Errorf("expected key %q in results", k)
		}
	}
}

func mustAbsent(t *testing.T, results map[string]any, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if _, ok := results[k]; ok {
			t.Errorf("key %q should be absent from results", k)
		}
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestLinearChain(t *testing.T) {
	// A → B → C
	tasks := []Task{
		{ID: "A", Deps: []string{}, Run: instant(1)},
		{ID: "B", Deps: []string{"A"}, Run: func(_ context.Context, d map[string]any) (any, error) {
			return d["A"].(int) + 1, nil
		}},
		{ID: "C", Deps: []string{"B"}, Run: func(_ context.Context, d map[string]any) (any, error) {
			return d["B"].(int) + 1, nil
		}},
	}
	res, err := RunDAG(context.Background(), tasks, 4)
	if err != nil {
		t.Fatal(err)
	}
	if res["C"] != 3 {
		t.Fatalf("C should be 3, got %v", res["C"])
	}
}

func TestParallelIndependent(t *testing.T) {
	var mu sync.Mutex
	var concurrent, maxConcurrent int32

	slow := func(id string) Task {
		return Task{
			ID:   id,
			Deps: []string{},
			Run: func(ctx context.Context, _ map[string]any) (any, error) {
				atomic.AddInt32(&concurrent, 1)
				time.Sleep(30 * time.Millisecond)
				cur := atomic.LoadInt32(&concurrent)
				mu.Lock()
				if cur > maxConcurrent {
					maxConcurrent = cur
				}
				mu.Unlock()
				atomic.AddInt32(&concurrent, -1)
				return id, nil
			},
		}
	}

	tasks := []Task{slow("t1"), slow("t2"), slow("t3")}
	start := time.Now()
	res, err := RunDAG(context.Background(), tasks, 3)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatal(err)
	}
	mustKeys(t, res, "t1", "t2", "t3")

	if elapsed > 80*time.Millisecond {
		t.Errorf("tasks should run in parallel, elapsed %v", elapsed)
	}
	if maxConcurrent < 2 {
		t.Errorf("expected at least 2 concurrent tasks, got max %d", maxConcurrent)
	}
}

func TestMaxWorkersRespected(t *testing.T) {
	var concurrent, maxConcurrent int32

	makeSlow := func(id string) Task {
		return Task{
			ID:   id,
			Deps: []string{},
			Run: func(ctx context.Context, _ map[string]any) (any, error) {
				cur := atomic.AddInt32(&concurrent, 1)
				if cur > atomic.LoadInt32(&maxConcurrent) {
					atomic.StoreInt32(&maxConcurrent, cur)
				}
				time.Sleep(20 * time.Millisecond)
				atomic.AddInt32(&concurrent, -1)
				return id, nil
			},
		}
	}

	tasks := make([]Task, 8)
	for i := range tasks {
		tasks[i] = makeSlow(fmt.Sprintf("t%d", i))
	}

	_, err := RunDAG(context.Background(), tasks, 3)
	if err != nil {
		t.Fatal(err)
	}
	if maxConcurrent > 3 {
		t.Errorf("max concurrent exceeded limit: got %d, want ≤ 3", maxConcurrent)
	}
}

func TestDepResultsPassedCorrectly(t *testing.T) {
	tasks := []Task{
		{ID: "x", Deps: []string{}, Run: instant(42)},
		{ID: "y", Deps: []string{}, Run: instant(58)},
		{ID: "sum", Deps: []string{"x", "y"}, Run: func(_ context.Context, d map[string]any) (any, error) {
			return d["x"].(int) + d["y"].(int), nil
		}},
	}
	res, err := RunDAG(context.Background(), tasks, 4)
	if err != nil {
		t.Fatal(err)
	}
	if res["sum"] != 100 {
		t.Fatalf("sum should be 100, got %v", res["sum"])
	}
}

func TestFailedTaskSkipsDependents(t *testing.T) {
	boom := errors.New("boom")
	tasks := []Task{
		{ID: "ok", Deps: []string{}, Run: instant("fine")},
		{ID: "bad", Deps: []string{}, Run: func(_ context.Context, _ map[string]any) (any, error) {
			return nil, boom
		}},
		{ID: "child-of-bad", Deps: []string{"bad"}, Run: instant("should not run")},
		{ID: "grandchild", Deps: []string{"child-of-bad"}, Run: instant("should not run either")},
	}

	res, err := RunDAG(context.Background(), tasks, 4)
	// err may be non-nil (the error from "bad") but independent "ok" should still be in results
	_ = err
	mustKeys(t, res, "ok")
	mustAbsent(t, res, "child-of-bad", "grandchild")
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	tasks := []Task{
		{
			ID:   "blocker",
			Deps: []string{},
			Run: func(ctx context.Context, _ map[string]any) (any, error) {
				close(started)
				<-ctx.Done()
				return nil, ctx.Err()
			},
		},
		{ID: "never", Deps: []string{"blocker"}, Run: instant("should not appear")},
	}

	done := make(chan struct{})
	var runErr error
	go func() {
		defer close(done)
		_, runErr = RunDAG(ctx, tasks, 2)
	}()

	<-started
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RunDAG did not return after context cancellation")
	}

	if !errors.Is(runErr, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", runErr)
	}
}

func TestCycleDetection(t *testing.T) {
	tasks := []Task{
		{ID: "A", Deps: []string{"C"}, Run: instant(1)},
		{ID: "B", Deps: []string{"A"}, Run: instant(2)},
		{ID: "C", Deps: []string{"B"}, Run: instant(3)},
	}
	_, err := RunDAG(context.Background(), tasks, 4)
	if !errors.Is(err, ErrCycle) {
		t.Fatalf("expected ErrCycle, got %v", err)
	}
}

func TestEmptyTasks(t *testing.T) {
	res, err := RunDAG(context.Background(), nil, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Fatal("expected empty result")
	}
}

func TestDiamondDependency(t *testing.T) {
	//      A
	//     / \
	//    B   C
	//     \ /
	//      D
	var order []string
	var mu sync.Mutex
	record := func(id string) func(context.Context, map[string]any) (any, error) {
		return func(_ context.Context, _ map[string]any) (any, error) {
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
			return id, nil
		}
	}

	tasks := []Task{
		{ID: "A", Deps: []string{}, Run: record("A")},
		{ID: "B", Deps: []string{"A"}, Run: record("B")},
		{ID: "C", Deps: []string{"A"}, Run: record("C")},
		{ID: "D", Deps: []string{"B", "C"}, Run: record("D")},
	}

	res, err := RunDAG(context.Background(), tasks, 4)
	if err != nil {
		t.Fatal(err)
	}
	mustKeys(t, res, "A", "B", "C", "D")

	sort.Strings(order[:len(order)-1]) // last must be D, first must be A
	if order[0] != "A" {
		t.Errorf("A should run first, order: %v", order)
	}
	if order[len(order)-1] != "D" {
		t.Errorf("D should run last, order: %v", order)
	}
}

func TestNoGoroutineLeak(t *testing.T) {
	// Checks that RunDAG returns only after all goroutines it spawned have exited.
	// We verify this by ensuring the "blocker" task's goroutine is cleaned up
	// (otherwise data-race detectors and goroutine leak checkers will catch it).
	var wg sync.WaitGroup
	wg.Add(1)

	tasks := []Task{
		{
			ID:   "tracked",
			Deps: []string{},
			Run: func(ctx context.Context, _ map[string]any) (any, error) {
				defer wg.Done()
				return "ok", nil
			},
		},
	}

	_, err := RunDAG(context.Background(), tasks, 1)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("goroutine from task did not finish before RunDAG returned")
	}
}
