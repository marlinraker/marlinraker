package executors

import (
	"net/http"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type MachineShutdownResult string

func MachineShutdown(*connections.Connection, *http.Request, Params) (any, error) {
	go func() {
		connections.TerminateAllConnections()
		if out, err := exec.Command("systemctl", "poweroff").CombinedOutput(); err != nil {
			log.Errorf("Failed to shutdown machine: %v", err)
			log.Errorln(string(out))
		}
	}()
	return "ok", nil
}
