package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

func PrinterPrintStart(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	if marlinraker.Printer == nil {
		return nil, util.NewError("printer is not online", 500)
	}

	fileName, err := params.RequireString("filename")
	if err != nil {
		return nil, err
	}

	if err := marlinraker.Printer.PrintManager.SelectFile(fileName); err != nil {
		return nil, err
	}

	if err := marlinraker.Printer.PrintManager.Start(); err != nil {
		return nil, err

	}
	return "ok", nil
}
