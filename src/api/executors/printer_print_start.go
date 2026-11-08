package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
	"strconv"
)

func PrinterPrintStart(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	printer := marlinraker.GetPrinter()
	if printer == nil {
		return nil, util.NewError(500, "printer is not online")
	}

	fileName, err := params.RequireString("filename")
	if err != nil {
		return nil, err
	}

	<-printer.MainExecutorContext().QueueGcode("SDCARD_PRINT_FILE FILENAME="+strconv.Quote(fileName), true)
	return "ok", nil
}
