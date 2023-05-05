package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/system_info"
	"net/http"
)

type MachineProcStatsResult = system_info.ProcStats

func MachineProcStats(*connections.Connection, *http.Request, Params) (any, error) {
	return system_info.GetStats(), nil
}
