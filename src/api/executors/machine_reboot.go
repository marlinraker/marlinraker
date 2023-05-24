package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
	"os/exec"
)

type MachineRebootResult string

func MachineReboot(*connections.Connection, *http.Request, Params) (any, error) {
	go func() {
		connections.TerminateAllConnections()
		if err := exec.Command("systemctl", "reboot").Err; err != nil {
			util.LogError(err)
		}
	}()
	return "ok", nil
}
