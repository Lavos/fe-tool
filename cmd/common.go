package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
)

func WaitForSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Block until a signal is received.
	fmt.Fprintf(os.Stderr, "Got signal `%s`, unblocking.\n", <-sig)
}
