package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

func PrinterFirmwareRestart(*connections.Connection, *http.Request, Params) (any, error) {
	if err := marlinraker.Reconnect(); err != nil {
		return nil, err
	}
	return "ok", nil
}
