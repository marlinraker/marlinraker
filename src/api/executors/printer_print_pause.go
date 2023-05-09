package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

func PrinterPrintPause(*connections.Connection, *http.Request, Params) (any, error) {
	if marlinraker.Printer == nil {
		return nil, util.NewError("printer is not online", 500)
	}
	if err := marlinraker.Printer.PrintManager.Pause(); err != nil {
		return nil, err
	}
	return "ok", nil
}
