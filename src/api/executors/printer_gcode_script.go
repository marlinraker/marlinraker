package executors

import (
	"fmt"
	"marlinraker-go/src/marlinraker"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/util"
	"net/http"
)

func PrinterGcodeScript(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	script, exists := params["script"]
	if !exists {
		return nil, util.NewError("script param is required", 400)
	}
	if marlinraker.Printer == nil {
		return nil, util.NewError("printer not online", 500)
	}
	ch, err := marlinraker.Printer.QueueGcode(fmt.Sprintf("%v", script), false, false)
	if err != nil {
		return nil, err
	}
	<-ch
	return "ok", err
}
