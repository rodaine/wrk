package wrk

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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

		out := w.Run(context.Background())

		var ne NamedError
		assert.ErrorAs(t, out, &ne)
		assert.Equal(t, w.Name, ne.Name)
		assert.ErrorIs(t, ne.Err, err)
	})

	t.Run("no error", func(t *testing.T) {
		t.Parallel()

		del := MockWorker(make(chan error, 1))
		del <- nil

		w := Named{
			Name:     "foo",
			Delegate: del,
		}

		out := w.Run(context.Background())
		assert.NoError(t, out)
	})
}

func TestNamed_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stop error", func(t *testing.T) {

		err := errors.New("some error")
		stopErr := make(chan error, 1)
		stopErr <- err

		w := Named{
			Name:     "foo",
			Delegate: MockWorkStopper{StopErr: stopErr},
		}

		out := w.Stop()

		var ne NamedError
		assert.ErrorAs(t, out, &ne)
		assert.Equal(t, w.Name, ne.Name)
		assert.ErrorIs(t, ne.Err, err)
		assert.Contains(t, ne.Error(), "foo")
		assert.Equal(t, err, ne.Unwrap())
	})

	t.Run("no stop", func(t *testing.T) {
		t.Parallel()

		w := Named{Name: "foo", Delegate: MockWorker(nil)}
		err := w.Stop()
		assert.NoError(t, err)
	})
}
