package executors

import (
	"net/http"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

func ServerRestart(*connections.Connection, *http.Request, Params) (any, error) {
	go func() {
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
			log.Errorf("Failed to restart server: %v", err)
		}
	}()
	return "ok", nil
}
