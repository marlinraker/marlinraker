package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/system_info"
)

type MachineSystemInfoResult struct {
	SystemInfo *system_info.SystemInfo `json:"system_info"`
}

func MachineSystemInfo(*connections.Connection, *http.Request, Params) (any, error) {

	info, err := system_info.GetSystemInfo()
	if err != nil {
		return nil, err
	}

	return MachineSystemInfoResult{info}, nil
}
