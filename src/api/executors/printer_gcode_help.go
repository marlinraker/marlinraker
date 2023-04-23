// TODO

package executors

import "marlinraker-go/src/marlinraker/connections"

type PrinterGcodeHelpResult map[string]string

func PrinterGcodeHelp(*connections.Connection, Params) (any, error) {
	return make(PrinterGcodeHelpResult), nil
}
