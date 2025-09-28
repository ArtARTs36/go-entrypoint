package entrypoint

import (
	"os"
	"os/signal"
)

type InterruptionListener func(chan os.Signal)

func OsSignalInterruptionListener(signals ...os.Signal) InterruptionListener {
	return func(done chan os.Signal) {
		signal.Notify(done, signals...)
	}
}
