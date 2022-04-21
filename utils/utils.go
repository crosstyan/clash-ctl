package utils

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cespare/xxhash/v2"
)

func Signal() chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	return sigCh
}

func GenHashString(s string) string {
	h := xxhash.Sum64String(s)
	return fmt.Sprintf("%x", h)
}
