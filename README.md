# go-entrypoint

```shell
go get github.com/artarts36/go-entrypoint
```

A go library for manage entrypoints like http/grpc server and graceful shutdown by signal.

## Usage

```go
package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/artarts36/go-entrypoint"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:              ":8001",
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	runner := entrypoint.NewRunner([]entrypoint.Entrypoint{
		entrypoint.HTTPServer("metrics", metricsServer),
	})

	if err := runner.Run(); err != nil {
		slog.Error("failed to run", slog.Any("err", err))
	}
}
```
