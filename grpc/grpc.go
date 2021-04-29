package grpc

import (
	"context"
	"errors"
	"net"
	"time"

	"google.golang.org/grpc"
)

const DefaultStopTimeout = 5 * time.Second

var noServerErr = errors.New("gRPC server must be set")

// Server is a wrk.WorkStopper responsible for running a grpc.Server, stopping
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

func (srv Server) Run(context.Context) error {
	if srv.Server == nil {
		return noServerErr
	}

	lis, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}

	return srv.Server.Serve(lis)
}

// Stop satisfies the wrk.WorkStopper interface. Note that if this method
// returns an error, the gRPC server may still be running.
func (srv Server) Stop() error {
	if srv.Server == nil {
		return noServerErr
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		srv.Server.GracefulStop()
	}()

	timeout := srv.StopTimeout
	if timeout == 0 {
		timeout = DefaultStopTimeout
	}

	select {
	case <-time.After(timeout):
		return context.DeadlineExceeded
	case <-done:
		return nil
	}
}
