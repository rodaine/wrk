package wrk

import (
	"context"
	"fmt"
)

// Named is a decorator WorkStopper that wraps any errors returned by Run or
// Stop as a NamedError.
type Named struct {
	// Name is the identifier for the wrapped Worker.
	Name string

	// Delegate is the Worker (or WorkStopper) associated with Name.
	Delegate Worker
}

// Run satisfies the [Worker] interface.
func (n Named) Run(ctx context.Context) error {
	return n.wrap(n.Delegate.Run(ctx))
}

// Stop satisfies the [WorkStopper] interface.
func (n Named) Stop() error {
	if ws, ok := n.Delegate.(WorkStopper); ok {
		return n.wrap(ws.Stop())
	}
	return nil
}

func (n Named) wrap(err error) error {
	if err == nil {
		return nil
	}
	return NamedError{
		Name: n.Name,
		Err:  err,
	}
}

// NamedError is the wrapped error returned by Named.
type NamedError struct {
	Name string
	Err  error
}

func (err NamedError) Error() string {
	return fmt.Sprintf("%s: %v", err.Name, err.Err)
}

func (err NamedError) Unwrap() error { return err.Err }

var (
	_ WorkStopper = Named{}
	_ error       = NamedError{}
)
