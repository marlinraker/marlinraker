package executors

import (
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/marlinraker/temp_store"
)

type ServerTemperatureStoreResult temp_store.TempStore

func ServerTemperatureStore(*connections.Connection, Params) (any, error) {
	return temp_store.GetStore(), nil
}
