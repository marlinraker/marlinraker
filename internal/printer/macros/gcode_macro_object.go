package macros

import "github.com/marlinraker/marlinraker/internal/printer_objects"

type gcodeMacroObject struct {
	variables map[string]any
}

func (object gcodeMacroObject) Query() (printer_objects.QueryResult, error) {
	return object.variables, nil
}
