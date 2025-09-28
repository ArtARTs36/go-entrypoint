package entrypoint

import (
	"context"
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

			return server.ListenAndServe()
		},
		Stop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	}
}
