package executors

import (
	"net/http"

	"github.com/samber/lo"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/printer_objects"
)

type PrinterObjectsListResult struct {
	Objects []string `json:"objects"`
}

func PrinterObjectsList(*connections.Connection, *http.Request, Params) (any, error) {
	return PrinterObjectsListResult{
		Objects: lo.Keys(printer_objects.GetObjects()),
	}, nil
}
