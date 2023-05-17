package printer

import (
	"marlinraker/src/printer_objects"
	"strings"
)

type toolheadObject struct {
	printer *Printer
}

func (object toolheadObject) Query() printer_objects.QueryResult {

	var homedAxes strings.Builder
	for i, homed := range object.printer.GcodeState.HomedAxes {
		if homed {
			homedAxes.WriteByte("xyz"[i])
		}
	}

	return printer_objects.QueryResult{
		"homed_axes":             homedAxes.String(),
		"axis_minimum":           append(object.printer.config.Printer.AxisMinimum[:], 0),
		"axis_maximum":           append(object.printer.config.Printer.AxisMaximum[:], 0),
		"print_time":             0,
		"stalls":                 0,
		"estimated_print_time":   0,
		"extruder":               "extruder",
		"position":               object.printer.GcodeState.Position,
		"max_velocity":           300,
		"max_accel":              3000,
		"max_accel_to_decel":     1500,
		"square_corner_velocity": 5,
	}
}
