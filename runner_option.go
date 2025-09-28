package entrypoint

import (
	"os"
	"syscall"
	"time"
)

type RunnerOption func(*runnerConfig)

type runnerConfig struct {
	shutdownTimeout      time.Duration
	interruptionListener InterruptionListener
}

func buildConfig(opts []RunnerOption) *runnerConfig {
	const defaultShutdownTimeout = 30 * time.Second

	cfg := &runnerConfig{
		shutdownTimeout:      defaultShutdownTimeout,
		interruptionListener: OsSignalInterruptionListener(os.Interrupt, syscall.SIGINT, syscall.SIGTERM),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func WithShutdownTimeout(timeout time.Duration) RunnerOption {
	return func(cfg *runnerConfig) {
		cfg.shutdownTimeout = timeout
	}
}

func WithInterruptionSignals(signals ...os.Signal) RunnerOption {
	return WithInterruptionListener(OsSignalInterruptionListener(signals...))
}

func WithInterruptionListener(listener InterruptionListener) RunnerOption {
	return func(cfg *runnerConfig) {
		cfg.interruptionListener = listener
	}
}
