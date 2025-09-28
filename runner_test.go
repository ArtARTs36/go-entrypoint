package entrypoint

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	interruptionListener := &mockInterruptionListener{}

	runner := NewRunner([]Entrypoint{
		{
			Name: "handler1",
			Run: func(ctx context.Context) error {
				return sleep(ctx, 10*time.Second)
			},
		},
		{
			Name: "handler2",
			Run: func(ctx context.Context) error {
				return sleep(ctx, 10*time.Second)
			},
		},
	}, WithInterruptionListener(interruptionListener.listen))

	go func() {
		err := runner.Run()
		require.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)

	started := time.Now()

	interruptionListener.notify(t)

	time.Sleep(100 * time.Millisecond)

	execution := time.Since(started)

	require.Less(t, execution, time.Second)
}

type mockInterruptionListener struct {
	ch chan os.Signal
}

func (l *mockInterruptionListener) listen(ch chan os.Signal) {
	l.ch = ch
}

func (l *mockInterruptionListener) notify(t *testing.T) {
	require.NotNil(t, l.ch, "signal channel not provided")

	l.ch <- os.Interrupt
}

func sleep(ctx context.Context, duration time.Duration) error {
	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
