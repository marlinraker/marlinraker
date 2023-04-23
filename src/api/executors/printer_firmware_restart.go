package executors

import (
	"marlinraker-go/src/marlinraker"
	"marlinraker-go/src/marlinraker/connections"
)

func PrinterFirmwareRestart(*connections.Connection, Params) (any, error) {
	if marlinraker.Printer != nil {
		if err := marlinraker.Printer.Disconnect(); err != nil {
			return nil, err
		}
	}
	go marlinraker.Connect()
	return "ok", nil
}
