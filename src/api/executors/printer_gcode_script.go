package executors

import (
	"fmt"
	"marlinraker-go/src/marlinraker"
	"marlinraker-go/src/marlinraker/connections"
)

func PrinterGcodeScript(_ *connections.Connection, params Params) (any, error) {
	script, exists := params["script"]
	if !exists {
		return nil, NewError("script param is required", 400)
	}
	if marlinraker.Printer == nil {
		return nil, NewError("printer not online", 500)
	}
	ch, err := marlinraker.Printer.QueueGcode(fmt.Sprintf("%v", script), false, false)
	if err != nil {
		return nil, err
	}
	<-ch
	return "ok", err
}
