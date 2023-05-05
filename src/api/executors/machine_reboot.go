package executors

import (
	"marlinraker/src/marlinraker/connections"
	"net/http"
	"syscall"
)

type MachineRebootResult struct{}

func MachineReboot(*connections.Connection, *http.Request, Params) (any, error) {
	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART); err != nil {
		return nil, err
	}
	return MachineRebootResult{}, nil
}
