package executors

import (
	"marlinraker-go/src/constants"
	"marlinraker-go/src/marlinraker"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/system_info"
	"os"
)

type PrinterInfoResult struct {
	State           marlinraker.KlippyState `json:"state"`
	StateMessage    string                  `json:"state_message"`
	Hostname        string                  `json:"hostname"`
	SoftwareVersion string                  `json:"software_version"`
	CpuInfo         string                  `json:"cpu_info"`
	KlipperPath     string                  `json:"klipper_path"`
	PythonPath      string                  `json:"python_path"`
	LogFile         string                  `json:"log_file"`
	ConfigFile      string                  `json:"config_file"`
}

func PrinterInfo(*connections.Connection, Params) (any, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	systemInfo, err := system_info.GetSystemInfo()
	if err != nil {
		return nil, err
	}

	return PrinterInfoResult{
		State:           marlinraker.State,
		StateMessage:    marlinraker.StateMessage,
		Hostname:        hostname,
		SoftwareVersion: constants.Version,
		CpuInfo:         systemInfo.CpuInfo.CpuDesc,
	}, nil
}
