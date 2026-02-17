package wrk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWork(t *testing.T) {
	t.Parallel()

	noErrs := make(chan error)
	close(noErrs)

	t.Run("nothing", func(t *testing.T) {
		t.Parallel()
		err := Work(t.Context())
		assert.NoError(t, err)
	})

	t.Run("clean exit", func(t *testing.T) {
		t.Parallel()

		mw := MockWorker(noErrs)
		w1 := MockWorkStopper{MockWorker: mw, StopErr: noErrs}
		w2 := MockWorkStopper{MockWorker: mw, StopErr: noErrs}

		err := Work(t.Context(), w1, w2)
		assert.NoError(t, err)
	})

	t.Run("context cancelled", func(t *testing.T) {
		t.Parallel()

		w1 := MockWorkStopper{StopErr: noErrs}
		w2 := MockWorkStopper{StopErr: noErrs}

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := Work(ctx, w1, w2)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("stop after return", func(t *testing.T) {
		t.Parallel()

		w1 := MockWorkStopper{StopErr: noErrs}
		w2 := MockWorkStopper{
			MockWorker: make(chan error),
			StopErr:    make(chan error),
		}

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		done := make(chan error, 1)
		go func() { done <- Work(ctx, w1, w2) }()

		select {
		case w2.MockWorker <- nil:
			// expected
		case w2.StopErr <- nil:
			t.Fatal("w2 should not be trying to stop")
		case err := <-done:
			t.Fatalf("worker should not be done: %v", err)
		case <-time.After(time.Second):
			t.Fatal("w2 is not waiting to receive on any channels")
		}

		select {
		case w2.StopErr <- nil:
		// expected
		case err := <-done:
			t.Fatalf("worker should not be done: %v", err)
		case <-time.After(time.Second):
			t.Fatal("w2 is not stopping")
		}

		cancel()

		select {
		case err := <-done:
			require.ErrorIs(t, err, context.Canceled)
		case <-time.After(time.Second):
			t.Fatal("worker should have stopped")
		}
	})
}

type MockWorker chan error

type MockWorkStopper struct {
	MockWorker

	StopErr chan error
}

func (m MockWorker) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-m:
		return err
	}
}

func (m MockWorkStopper) Stop() error {
	return <-m.StopErr
}

var (
	_ Worker      = (MockWorker)(nil)
	_ WorkStopper = MockWorkStopper{}
)
