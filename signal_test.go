package wrk

import (
	"context"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignals_Run(t *testing.T) { //nolint:tparallel // subtests cannot be parallel as they are sending real signals
	t.Parallel()

	t.Run("captured", func(t *testing.T) { //nolint:paralleltest // cannot be parallel since we are sending a real signal.
		w := Signals{syscall.SIGHUP}
		ready := make(chan struct{})
		done := make(chan error, 1)

		go func() {
			close(ready)
			done <- w.Run(t.Context())
		}()

		<-ready

		p, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)
		err = p.Signal(syscall.SIGHUP)
		require.NoError(t, err)

		err = <-done
		var re ReceivedSignalError
		require.ErrorAs(t, err, &re)
		assert.Equal(t, syscall.SIGHUP, re.Signal)
		assert.Contains(t, re.Error(), syscall.SIGHUP.String())
	})

	t.Run("canceled", func(t *testing.T) { //nolint:paralleltest // cannot be parallel since we are sending a real signal.
		w := Signals{syscall.SIGHUP}

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := w.Run(ctx)
		require.ErrorIs(t, err, context.Canceled)
	})
}
