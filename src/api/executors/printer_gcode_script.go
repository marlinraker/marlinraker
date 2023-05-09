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

	if marlinraker.Printer == nil {
		return nil, util.NewError("printer is not online", 500)
	}
	ch, err := marlinraker.Printer.QueueGcode(script, false, false)
	if err != nil {
		return nil, err
	}
	<-ch
	return "ok", err
}
