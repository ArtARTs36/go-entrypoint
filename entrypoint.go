package entrypoint

import "context"

type Entrypoint struct {
	// Required.
	Name string

	// Required.
	Run func(ctx context.Context) error
}
