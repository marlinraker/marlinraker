package executors

import (
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/system_info"
)

type MachineSystemInfoResult struct {
	SystemInfo *system_info.SystemInfo `json:"system_info"`
}

func MachineSystemInfo(*connections.Connection, Params) (any, error) {

	info, err := system_info.GetSystemInfo()
	if err != nil {
		return nil, err
	}

	return MachineSystemInfoResult{info}, nil
}
