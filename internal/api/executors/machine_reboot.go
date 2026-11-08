package executors

import (
	"net/http"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type MachineRebootResult string

func MachineReboot(*connections.Connection, *http.Request, Params) (any, error) {
	go func() {
		connections.TerminateAllConnections()
		if out, err := exec.Command("systemctl", "reboot").CombinedOutput(); err != nil {
			log.Errorf("Failed to reboot machine: %v", err)
			log.Errorln(string(out))
		}
	}()
	return "ok", nil
}
