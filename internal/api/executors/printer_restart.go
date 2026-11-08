package executors

import (
	"net/http"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

func PrinterRestart(*connections.Connection, *http.Request, Params) (any, error) {
	go func() {
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
			log.Errorf("Failed to restart printer: %v", err)
		}
	}()
	return "ok", nil
}
