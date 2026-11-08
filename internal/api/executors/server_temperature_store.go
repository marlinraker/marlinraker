package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/marlinraker/temp_store"
)

type ServerTemperatureStoreResult temp_store.TempStore

func ServerTemperatureStore(*connections.Connection, *http.Request, Params) (any, error) {
	return temp_store.GetStore(), nil
}
