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
	<-marlinraker.Printer.MainExecutorContext().QueueGcode("RESUME", true)
	return "ok", nil
}
