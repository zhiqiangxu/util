package signal

import (
	"os"
	"os/signal"
)

// SetupHandler will call handler when signal received
func SetupHandler(handler func(os.Signal), s ...os.Signal) {
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, s...)

	go func() {
		for sig := range sigCh {
			handler(sig)
		}
	}()
}
