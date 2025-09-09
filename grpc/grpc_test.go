package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestServer(t *testing.T) {
	t.Parallel()

	t.Run("zero", func(t *testing.T) {
		t.Parallel()

		w := &Server{}
		err := w.Run(t.Context())
		require.ErrorIs(t, err, errNoServer)
		err = w.Stop()
		require.ErrorIs(t, err, errNoServer)
	})

	t.Run("server", func(t *testing.T) {
		t.Parallel()

		addr := "localhost:11111"
		srv := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
		healthSvc := health.NewServer()
		grpc_health_v1.RegisterHealthServer(srv, healthSvc)
		healthSvc.SetServingStatus("foo", grpc_health_v1.HealthCheckResponse_SERVING)

		w := &Server{
			Server: srv,
			Addr:   addr,
		}

		done := make(chan error, 1)
		go func() {
			done <- w.Run(t.Context())
		}()
		defer func() {
			err := w.Stop()
			require.NoError(t, err)
			require.NoError(t, <-done)
		}()

		conn, err := grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()
		resp, err := grpc_health_v1.NewHealthClient(conn).Check(
			t.Context(),
			&grpc_health_v1.HealthCheckRequest{Service: "foo"},
		)
		require.NoError(t, err)
		require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.GetStatus())
	})

	t.Run("stop_without_start", func(t *testing.T) {
		t.Parallel()
		w := &Server{
			Server: grpc.NewServer(grpc.Creds(insecure.NewCredentials())),
			Addr:   ":0",
		}
		require.NoError(t, w.Stop())
	})

	t.Run("addr error", func(t *testing.T) {
		t.Parallel()
		w := &Server{
			Server: grpc.NewServer(grpc.Creds(insecure.NewCredentials())),
			Addr:   "foo:bar",
		}
		require.Error(t, w.Run(t.Context()))
	})
}
