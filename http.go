package wrk

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

const DefaultHTTPStopTimeout = 5 * time.Second

// HTTPServer executes an http.Server, stopping it gracefully via its Shutdown
// method.
type HTTPServer struct {
	// Server is the http.Server instance that will be run. If nil, the default,
	// zero-value server will be executed instead. See the http.Server
	// documentation details on the default behavior.
	Server *http.Server

	// StopTimeout is the grace-period allowed for the graceful shutdown of the
	// Server to complete. If zero, the DefaultHTTPStopTimeout value is used.
	StopTimeout time.Duration

	// If OverrideBaseContext is true, the base context attached to the
	// http.Request's handled by the Server are replaced with the context
	// passed to Run. Note, this means all in-flight requests will have their
	// context canceled during graceful shutdown
	OverrideBaseContext bool

	once sync.Once
}

func (srv *HTTPServer) init() {
	srv.once.Do(func() {
		if srv.Server == nil {
			srv.Server = &http.Server{}
		}
	})
}

func (srv *HTTPServer) Run(ctx context.Context) error {
	srv.init()

	if srv.OverrideBaseContext {
		srv.Server.BaseContext = func(net.Listener) context.Context {
			return ctx
		}
	}

	err := srv.Server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (srv *HTTPServer) Stop() error {
	srv.init()

	timeout := srv.StopTimeout
	if timeout == 0 {
		timeout = DefaultHTTPStopTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return srv.Server.Shutdown(ctx)
}

var _ WorkStopper = (*HTTPServer)(nil)
