package executors

import (
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/system_info"
)

type MachineProcStatsResult = system_info.ProcStats

func MachineProcStats(*connections.Connection, Params) (any, error) {
	return system_info.GetStats(), nil
}
