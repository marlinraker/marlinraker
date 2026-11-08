package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/util"
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
