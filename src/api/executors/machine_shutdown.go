package executors

import (
	"marlinraker/src/marlinraker/connections"
	"net/http"
	"syscall"
)

type MachineShutdownResult struct{}

func MachineShutdown(*connections.Connection, *http.Request, Params) (any, error) {
	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		return nil, err
	}
	return MachineShutdownResult{}, nil
}
