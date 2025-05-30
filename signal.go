package wrk

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

// Signals is a Worker that blocks until a specified os.Signal is received or
// the context is canceled. This worker always returns an error, either a
// ReceivedSignalError or the context's error when canceled. After the first
// signal is received, any subsequent signals will not be captured and will
// follow their default behavior.
//
// See signal.Notify for documentation on the signal capture behavior.
type Signals []os.Signal

// Run satisfies the [Worker] interface.
func (s Signals) Run(ctx context.Context) error {
	notify := make(chan os.Signal, 1)
	signal.Notify(notify, s...)
	defer signal.Stop(notify)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case sig := <-notify:
		return ReceivedSignalError{sig}
	}
}

// ReceivedSignalError is returned by Signals.Run when a signal is captured.
type ReceivedSignalError struct {
	// Signal is the captured os.Signal
	Signal os.Signal
}

func (err ReceivedSignalError) Error() string {
	return fmt.Sprint("received signal: ", err.Signal)
}

var (
	_ Worker = Signals{}
	_ error  = ReceivedSignalError{}
)
