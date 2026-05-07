package entrypoint

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

type Runner struct {
	entrypoints          []Entrypoint
	shutdownTimeout      time.Duration
	interruptionListener InterruptionListener
}

func Run(es []Entrypoint, option ...RunnerOption) error {
	return NewRunner(es, option...).Run()
}

func NewRunner(
	es []Entrypoint,
	option ...RunnerOption,
) *Runner {
	if len(es) == 0 {
		panic("at least one entrypoint must be defined")
	}

	cfg := buildConfig(option)

	return &Runner{entrypoints: es, shutdownTimeout: cfg.shutdownTimeout, interruptionListener: cfg.interruptionListener}
}

func (b *Runner) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

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
		cancel()
	case err := <-wgChannel:
		slog.Error("[entrypoint] received error", slog.Any("err", err))
	}

	return nil
}
