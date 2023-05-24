package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
	"os/exec"
)

type MachineShutdownResult string

func MachineShutdown(*connections.Connection, *http.Request, Params) (any, error) {
	go func() {
		connections.TerminateAllConnections()
		if err := exec.Command("systemctl", "poweroff").Err; err != nil {
			util.LogError(err)
		}
	}()
	return "ok", nil
}
