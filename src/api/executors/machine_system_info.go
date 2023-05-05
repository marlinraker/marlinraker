package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/system_info"
	"net/http"
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
