package job

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunAll(t *testing.T) {
	var n int32
	r := NewRunner()
	r.Add("a", func(ctx context.Context) error {
		atomic.AddInt32(&n, 1)
		return nil
	})
	r.Add("b", func(ctx context.Context) error {
		atomic.AddInt32(&n, 1)
		return nil
	})
	r.RunAll(context.Background())
	if atomic.LoadInt32(&n) != 2 {
		t.Fatalf("expected 2 runs, got %d", n)
	}
}

func TestRunPeriodicallyStopsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	runner := NewRunner()

	go func() {
		runner.RunPeriodically(ctx, time.Hour)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runner did not stop after context cancellation")
	}
}
