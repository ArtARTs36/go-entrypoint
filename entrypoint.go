package entrypoint

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"time"
)

type Bag struct {
	Entrypoints     []Entrypoint
	ShutdownTimeout time.Duration
}

type Entrypoint struct {
	Name string
	Run  func() error
	Stop func(ctx context.Context) error
}

func Wrap(es []Entrypoint, shutdownTimeout time.Duration) *Bag {
	return &Bag{Entrypoints: es, ShutdownTimeout: shutdownTimeout}
}

func (b *Bag) Run() error {
	wg, _ := errgroup.WithContext(context.Background())

	for _, e := range b.Entrypoints {
		wg.Go(func() error {
			err := e.Run()
			if err != nil {
				return fmt.Errorf("run %s: %w", e.Name, err)
			}
			return nil
		})
	}

	return wg.Wait()
}

func (b *Bag) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), b.ShutdownTimeout)
	defer cancel()

	for _, e := range b.Entrypoints {
		if err := e.Stop(ctx); err != nil {
			return fmt.Errorf("stop %q: %w", e.Name, err)
		}
	}

	return nil
}
