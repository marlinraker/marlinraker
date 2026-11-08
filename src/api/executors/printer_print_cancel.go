package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

func PrinterPrintCancel(*connections.Connection, *http.Request, Params) (any, error) {
	printer := marlinraker.GetPrinter()
	if printer == nil {
		return nil, util.NewError(500, "printer is not online")
	}
	<-printer.MainExecutorContext().QueueGcode("CANCEL_PRINT", true)
	return "ok", nil
}
