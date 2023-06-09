package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/marlinraker/gcode_store"
	"net/http"
)

type ServerGcodeStoreResult struct {
	GcodeStore []gcode_store.GcodeLog `json:"gcode_store"`
}

func ServerGcodeStore(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	count, exists := params.GetInt64("count")
	if !exists {
		count = 1000
	}

	start := len(gcode_store.GcodeStore) - int(count)
	if start < 0 {
		start = 0
	}

	return ServerGcodeStoreResult{
		GcodeStore: gcode_store.GcodeStore[start:],
	}, nil
}
