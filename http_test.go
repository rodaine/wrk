package wrk

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPServer(t *testing.T) {
	t.Parallel()

	t.Run("zero", func(t *testing.T) {
		t.Parallel()

		w := &HTTPServer{}
		done := make(chan error, 1)
		go func() {
			done <- w.Run(t.Context())
		}()

		err := w.Stop()
		assert.NoError(t, err)

		err = <-done
		assert.NoError(t, err)
	})

	t.Run("server", func(t *testing.T) {
		t.Parallel()

		addr := "localhost:12345"

		w := &HTTPServer{
			OverrideBaseContext: true,
			Server: &http.Server{ //nolint:gosec // fine for a test
				Addr: addr,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					val := req.Context().Value("foo")
					assert.Equal(t, "bar", val)
					w.WriteHeader(http.StatusTeapot)
				}),
			},
		}

		ctx := context.WithValue(t.Context(), "foo", "bar") //nolint:revive, staticcheck // fine for a test

		done := make(chan error, 1)
		go func() {
			done <- w.Run(ctx)
		}()
		defer func() {
			err := w.Stop()
			require.NoError(t, err)
			require.NoError(t, <-done)
		}()

		// Poll until the server is accepting connections.
		require.EventuallyWithT(t, func(ct *assert.CollectT) {
			res, err := http.Get("http://" + addr) //nolint:noctx // fine for a test
			if !assert.NoError(ct, err) {
				return
			}
			defer func() { _ = res.Body.Close() }()
			assert.Equal(ct, http.StatusTeapot, res.StatusCode)
		}, 2*time.Second, 10*time.Millisecond)
	})

	t.Run("stop without start", func(t *testing.T) {
		t.Parallel()
		w := &HTTPServer{Server: &http.Server{Addr: ":0"}} //nolint:gosec // fine for a test
		assert.NoError(t, w.Stop())
	})

	t.Run("listen error", func(t *testing.T) {
		t.Parallel()

		w := &HTTPServer{Server: &http.Server{Addr: "foo:bar"}} //nolint:gosec // test code is OK
		assert.Error(t, w.Run(t.Context()))
	})
}
