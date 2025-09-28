package entrypoint

import (
	"context"
	"errors"
	"net"
	"net/http"
)

func HTTPServer(name string, server *http.Server) Entrypoint {
	return Entrypoint{
		Name: name,
		Run: func(ctx context.Context) error {
			server.BaseContext = func(_ net.Listener) context.Context {
				return ctx
			}

			err := server.ListenAndServe()
			if err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return err
			}
			return nil
		},
		Stop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	}
}
