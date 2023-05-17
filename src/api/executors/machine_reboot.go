package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
	"syscall"
)

type MachineRebootResult string

func MachineReboot(*connections.Connection, *http.Request, Params) (any, error) {
	go func() {
		connections.TerminateAllConnections()
		if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART); err != nil {
			util.LogError(err)
		}
	}()
	return "ok", nil
}
