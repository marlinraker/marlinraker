package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

func PrinterFirmwareRestart(*connections.Connection, *http.Request, Params) (any, error) {
	if marlinraker.Printer != nil {
		if err := marlinraker.Printer.Disconnect(); err != nil {
			return nil, err
		}
	}
	go marlinraker.Connect()
	return "ok", nil
}
