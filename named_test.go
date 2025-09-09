package wrk

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamed_Run(t *testing.T) {
	t.Parallel()

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("some error")
		del := MockWorker(make(chan error, 1))
		del <- err

		w := Named{
			Name:     "foo",
			Delegate: del,
		}

		out := w.Run(t.Context())

		var ne NamedError
		require.ErrorAs(t, out, &ne)
		assert.Equal(t, w.Name, ne.Name)
		require.ErrorIs(t, ne.Err, err)
	})

	t.Run("no error", func(t *testing.T) {
		t.Parallel()

		del := MockWorker(make(chan error, 1))
		del <- nil

		w := Named{
			Name:     "foo",
			Delegate: del,
		}

		out := w.Run(t.Context())
		require.NoError(t, out)
	})
}

func TestNamed_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stop error", func(t *testing.T) {
		t.Parallel()
		err := errors.New("some error")
		stopErr := make(chan error, 1)
		stopErr <- err

		w := Named{
			Name:     "foo",
			Delegate: MockWorkStopper{StopErr: stopErr},
		}

		out := w.Stop()

		var ne NamedError
		require.ErrorAs(t, out, &ne)
		assert.Equal(t, w.Name, ne.Name)
		require.ErrorIs(t, ne.Err, err)
		assert.Contains(t, ne.Error(), "foo")
		assert.Equal(t, err, ne.Unwrap())
	})

	t.Run("no stop", func(t *testing.T) {
		t.Parallel()

		w := Named{Name: "foo", Delegate: MockWorker(nil)}
		err := w.Stop()
		require.NoError(t, err)
	})
}
