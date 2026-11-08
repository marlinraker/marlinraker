package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

func PrinterGcodeScript(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	script, err := params.RequireString("script")
	if err != nil {
		return nil, err
	}

	printer := marlinraker.GetPrinter()
	if printer == nil {
		return nil, util.NewError(500, "printer is not online")
	}
	<-printer.MainExecutorContext().QueueGcode(script, false)
	return "ok", err
}
