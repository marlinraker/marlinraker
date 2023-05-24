package executors

import (
	log "github.com/sirupsen/logrus"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
	"os/exec"
)

type MachineRebootResult string

func MachineReboot(*connections.Connection, *http.Request, Params) (any, error) {
	go func() {
		connections.TerminateAllConnections()
		if out, err := exec.Command("systemctl", "reboot").CombinedOutput(); err != nil {
			log.Error(string(out))
			util.LogError(err)
		}
	}()
	return "ok", nil
}
