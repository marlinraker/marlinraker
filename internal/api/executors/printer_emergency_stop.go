package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/util"
)

type PrinterEmergencyStopResult string

func PrinterEmergencyStop(*connections.Connection, *http.Request, Params) (any, error) {
	printer := marlinraker.GetPrinter()
	if printer == nil {
		return nil, util.NewError(500, "printer is not connected")
	}
	printer.EmergencyStop()
	return "ok", nil
}
