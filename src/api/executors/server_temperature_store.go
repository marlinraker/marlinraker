package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/marlinraker/temp_store"
	"net/http"
)

type ServerTemperatureStoreResult temp_store.TempStore

func ServerTemperatureStore(*connections.Connection, *http.Request, Params) (any, error) {
	return temp_store.GetStore(), nil
}
