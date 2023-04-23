package executors

import (
	"github.com/samber/lo"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/printer_objects"
)

type PrinterObjectsListResult struct {
	Objects []string `json:"objects"`
}

func PrinterObjectsList(*connections.Connection, Params) (any, error) {
	return PrinterObjectsListResult{
		Objects: lo.Keys(printer_objects.GetObjects()),
	}, nil
}
