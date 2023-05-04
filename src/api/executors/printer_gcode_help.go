// TODO

package executors

import (
	"marlinraker-go/src/marlinraker/connections"
	"net/http"
)

type PrinterGcodeHelpResult map[string]string

func PrinterGcodeHelp(*connections.Connection, *http.Request, Params) (any, error) {
	return make(PrinterGcodeHelpResult), nil
}
