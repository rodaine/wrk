package wrk

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Worker describes a long-running process that's executed concurrently with
// other workers within an application. Some examples of workers would be a web
// server (e.g., HTTPServer) or a daemon listening on a data stream.
type Worker interface {
	// Run should be a blocking operation that continues to execute until ctx
	// is canceled. Run should return an error only if the operation did not
	// exit intentionally (i.e., http.ErrServerClosed should be converted to
	// nil).
	//
	// If the work cannot be cancelled directly via the ctx (e.g.,
	// http.ListenAndServe) or resources need to be cleaned up out-of-band, a
	// Stop function should be provided. See WorkStopper for the appropriate
	// interface.
	Run(ctx context.Context) error
}

// WorkerFunc is a Worker defined as just by just its Run method.
type WorkerFunc func(ctx context.Context) error

func (w WorkerFunc) Run(ctx context.Context) error { return w(ctx) }

// A WorkStopper is a Worker that needs to be shutdown from an external
// goroutine.
type WorkStopper interface {
	Worker

	// Stop is called to signal that execution of the Worker should stop
	// (typically when another Worker returns an error or the root context is
	// canceled). Stop will always be called to allow any cleanup operations to
	// occur. Stop should return an error if the Worker cannot be shutdown
	// gracefully (or within a reasonable amount of time).
	//
	// Note: it is possible that both Stop and Run can be executed concurrently.
	// Any mutations from either method must be synchronized to avoid data races.
	Stop() error
}

func noopCancel() {}

// Work runs all the provided workers concurrently, terminating if any Worker
// returns an error or if the provided context is canceled. This function
// blocks until all workers have shutdown. The first error produced by a Worker
// is returned; it is nil if all workers have stopped cleanly.
func Work(ctx context.Context, workers ...Worker) error {
	grp, ctx := errgroup.WithContext(ctx)

	for _, w := range workers {
		wrk := w
		wCtx, wCancel := ctx, noopCancel

		if ws, ok := wrk.(WorkStopper); ok {
			wCtx, wCancel = context.WithCancel(ctx)

			grp.Go(func() error {
				<-wCtx.Done()
				return ws.Stop()
			})
		}

		grp.Go(func() error {
			defer wCancel()
			return wrk.Run(wCtx)
		})
	}

	return grp.Wait()
}
