package executors

import (
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

type PrinterGcodeHelpResult map[string]string

func PrinterGcodeHelp(*connections.Connection, *http.Request, Params) (any, error) {
	printer := marlinraker.GetPrinter()
	if printer == nil {
		return nil, util.NewError(500, "printer is not online")
	}

	help := make(PrinterGcodeHelpResult)
	for name, macro := range printer.MacroManager.Macros {
		help[name] = macro.Description()
	}
	return help, nil
}
