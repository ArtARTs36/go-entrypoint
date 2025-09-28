package entrypoint

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

type Runner struct {
	entrypoints          []Entrypoint
	shutdownTimeout      time.Duration
	interruptionListener InterruptionListener

	shuttingDown atomic.Bool
}

func NewRunner(
	es []Entrypoint,
	option ...RunnerOption,
) *Runner {
	if len(es) == 0 {
		panic("at least one entrypoint must be defined")
	}

	cfg := buildConfig(option)

	for i := range es {
		if es[i].Stop == nil {
			es[i].Stop = func(_ context.Context) error {
				return nil
			}
		}
	}

	return &Runner{entrypoints: es, shutdownTimeout: cfg.shutdownTimeout, interruptionListener: cfg.interruptionListener}
}

func (b *Runner) Run() error {
	wg, ctx := errgroup.WithContext(context.Background())

	done := make(chan os.Signal, 1)
	b.interruptionListener(done)

	for _, e := range b.entrypoints {
		wg.Go(func() error {
			slog.Info("[entrypoint] starting", slog.String("entrypoint.name", e.Name))

			err := e.Run(ctx)
			if err != nil {
				return fmt.Errorf("run %s: %w", e.Name, err)
			}
			return nil
		})
	}

	wgChannel := make(chan error, 1)

	go func() {
		var wgErr error

		defer func() {
			wgChannel <- wgErr
		}()

		wgErr = wg.Wait()
	}()

	select {
	case sig := <-done:
		slog.Info("[entrypoint] received shutdown signal", slog.String("signal", sig.String()))

		b.stop()
	case err := <-wgChannel:
		slog.Error("[entrypoint] received error", slog.Any("err", err))
		b.stop()
	}

	return nil
}

func (b *Runner) stop() {
	b.shuttingDown.Store(true)

	ctx, cancel := context.WithTimeout(context.Background(), b.shutdownTimeout)
	defer cancel()

	for _, e := range b.entrypoints {
		slog.Info("[entrypoint] stopping", slog.String("entrypoint.name", e.Name))

		if err := e.Stop(ctx); err != nil {
			slog.Info("[entrypoint] failed to stop", slog.String("entrypoint.name", e.Name), slog.Any("err", err))
		} else {
			slog.Info("[entrypoint] stopped", slog.String("entrypoint.name", e.Name), slog.Any("err", err))
		}
	}
}
