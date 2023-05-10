package macros

import "marlinraker/src/printer_objects"

type gcodeMacroObject struct {
	variables map[string]any
}

func (object gcodeMacroObject) Query() printer_objects.QueryResult {
	return object.variables
}
