package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

func PrinterFirmwareRestart(*connections.Connection, *http.Request, Params) (any, error) {
	if err := marlinraker.Reconnect(); err != nil {
		return nil, err
	}
	return "ok", nil
}
