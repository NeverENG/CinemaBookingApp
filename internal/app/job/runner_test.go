package job

import (
	"context"
	"sync/atomic"
	"testing"
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
