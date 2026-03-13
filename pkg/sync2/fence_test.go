package sync2

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestFence(t *testing.T) {
	t.Parallel()

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	var group errgroup.Group
	var fence Fence
	var done int32

	for i := 0; i < 10; i++ {
		group.Go(func() error {
			if !fence.Wait(ctx) {
				return errors.New("got false from Wait")
			}
			if atomic.LoadInt32(&done) == 0 {
				return errors.New("fence not yet released")
			}
			return nil
		})
	}

	// wait a bit for all goroutines to hit the fence
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 3; i++ {
		group.Go(func() error {
			atomic.StoreInt32(&done, 1)
			fence.Release()
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestFence_ContextCancel(t *testing.T) {
	t.Parallel()

	tctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx, cancel := context.WithCancel(tctx)

	var group errgroup.Group
	var fence Fence

	for i := 0; i < 10; i++ {
		group.Go(func() error {
			if fence.Wait(ctx) {
				return errors.New("got true from Wait")
			}
			return nil
		})
	}

	// wait a bit for all goroutines to hit the fence
	time.Sleep(100 * time.Millisecond)

	cancel()

	if err := group.Wait(); err != nil {
		t.Fatal(err)
	}
}
