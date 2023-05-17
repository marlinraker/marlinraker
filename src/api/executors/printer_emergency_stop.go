package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type PrinterEmergencyStopResult string

func PrinterEmergencyStop(*connections.Connection, *http.Request, Params) (any, error) {
	err := marlinraker.Printer.EmergencyStop()
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
