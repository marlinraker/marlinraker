package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/system_info"
)

type MachineProcStatsResult = system_info.ProcStats

func MachineProcStats(*connections.Connection, *http.Request, Params) (any, error) {
	return system_info.GetStats(), nil
}
