package wrk

import (
	"context"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignals_Run(t *testing.T) {
	t.Parallel()

	t.Run("captured", func(t *testing.T) {
		w := Signals{syscall.SIGHUP}
		ready := make(chan struct{})
		done := make(chan error, 1)

		go func() {
			close(ready)
			done <- w.Run(context.Background())
		}()

		<-ready

		p, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)
		err = p.Signal(syscall.SIGHUP)
		require.NoError(t, err)

		err = <-done
		var re ReceivedSignalError
		assert.ErrorAs(t, err, &re)
		assert.Equal(t, syscall.SIGHUP, re.Signal)
		assert.Contains(t, re.Error(), syscall.SIGHUP.String())
	})

	t.Run("canceled", func(t *testing.T) {
		w := Signals{syscall.SIGHUP}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := w.Run(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
