package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

func PrinterPrintResume(*connections.Connection, *http.Request, Params) (any, error) {
	if marlinraker.Printer == nil {
		return nil, util.NewError("printer is not online", 500)
	}
	if err := marlinraker.Printer.PrintManager.Resume(); err != nil {
		return nil, err
	}
	return "ok", nil
}
