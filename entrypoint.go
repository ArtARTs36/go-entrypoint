package entrypoint

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type Bag struct {
	Entrypoints     []Entrypoint
	ShutdownTimeout time.Duration
}

type Entrypoint struct {
	Name string
	Run  func(ctx context.Context) error
	Stop func(ctx context.Context) error
}

func Wrap(es []Entrypoint, shutdownTimeout time.Duration) *Bag {
	for i := range es {
		if es[i].Stop == nil {
			es[i].Stop = func(ctx context.Context) error {
				return nil
			}
		}
	}

	return &Bag{Entrypoints: es, ShutdownTimeout: shutdownTimeout}
}

func (b *Bag) Run() error {
	wg, ctx := errgroup.WithContext(context.Background())

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	for _, e := range b.Entrypoints {
		wg.Go(func() error {
			err := e.Run(ctx)
			if err != nil {
				return fmt.Errorf("run %s: %w", e.Name, err)
			}
			return nil
		})
	}

	<-done

	slog.Info("[entrypoint] received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), b.ShutdownTimeout)
	defer cancel()

	errs := []error{}

	for _, e := range b.Entrypoints {
		slog.Info("[entrypoint] stopping", slog.String("entrypoint.name", e.Name))

		if err := e.Stop(ctx); err != nil {
			errs = append(errs, err)

			slog.Info("[entrypoint] failed to stop", slog.String("entrypoint.name", e.Name))
		} else {
			slog.Info("[entrypoint] stopped", slog.String("entrypoint.name", e.Name))
		}
	}

	return errors.Join(errs...)
}
