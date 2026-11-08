package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

func PrinterPrintResume(*connections.Connection, *http.Request, Params) (any, error) {
	printer := marlinraker.GetPrinter()
	if printer == nil {
		return nil, util.NewError(500, "printer is not online")
	}
	<-printer.MainExecutorContext().QueueGcode("RESUME", true)
	return "ok", nil
}
