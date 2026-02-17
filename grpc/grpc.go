// Package grpc provides a [wrk.WorkStopper] implementation for gRPC servers.
package grpc

import (
	"cmp"
	"context"
	"errors"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/rodaine/wrk"
)

// DefaultStopTimeout is the default grace period used when stopping a [Server].
const DefaultStopTimeout = wrk.DefaultHTTPStopTimeout

var errNoServer = errors.New("gRPC server must be set")

// Server is a [wrk.WorkStopper] responsible for running a [grpc.Server], stopping
// it via its GracefulStop method.
type Server struct {
	// Server is the grpc.Server to run. This field must be non-nil.
	Server *grpc.Server

	// Addr is the "host:port" pair for the Server to listen on.
	Addr string

	// StopTimeout is the duration to wait for GracefulStop to complete. If
	// unset, DefaultStopTimeout is used.
	StopTimeout time.Duration
}

// Run satisfies the [wrk.Worker] interface.
func (srv Server) Run(ctx context.Context) error {
	if srv.Server == nil {
		return errNoServer
	}

	lis, err := new(net.ListenConfig).Listen(ctx, "tcp", srv.Addr)
	if err != nil {
		return err
	}

	return srv.Server.Serve(lis)
}

// Stop satisfies the [wrk.WorkStopper] interface. Note that if this method
// returns an error, the gRPC server may still be running.
func (srv Server) Stop() error {
	if srv.Server == nil {
		return errNoServer
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		srv.Server.GracefulStop()
	}()

	timeout := cmp.Or(srv.StopTimeout, DefaultStopTimeout)
	select {
	case <-time.After(timeout):
		return context.DeadlineExceeded
	case <-done:
		return nil
	}
}

var _ wrk.WorkStopper = (*Server)(nil)
