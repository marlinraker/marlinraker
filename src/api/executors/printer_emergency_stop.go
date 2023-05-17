package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type PrinterEmergencyStopResult string

func PrinterEmergencyStop(*connections.Connection, *http.Request, Params) (any, error) {
	marlinraker.Printer.EmergencyStop()
	return "ok", nil
}
