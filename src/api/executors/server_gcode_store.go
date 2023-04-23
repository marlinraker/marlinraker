package executors

import (
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/marlinraker/gcode_store"
)

type ServerGcodeStoreResult struct {
	GcodeStore []gcode_store.GcodeLog `json:"gcode_store"`
}

func ServerGcodeStore(_ *connections.Connection, params Params) (any, error) {

	count, exists := params["count"].(int)
	if !exists {
		count = 1000
	}

	start := len(gcode_store.GcodeStore) - count
	if start < 0 {
		start = 0
	}

	return ServerGcodeStoreResult{
		GcodeStore: gcode_store.GcodeStore[start:],
	}, nil
}
