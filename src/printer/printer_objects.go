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

type motionReportObject struct {
	printer *Printer
}

func (object motionReportObject) Query() printer_objects.QueryResult {
	return printer_objects.QueryResult{
		"live_position":          object.printer.GcodeState.Position,
		"live_velocity":          object.printer.GcodeState.Velocity,
		"live_extruder_velocity": object.printer.GcodeState.EVelocity,
	}
}

type gcodeMoveObject struct {
	printer *Printer
}

func (object gcodeMoveObject) Query() printer_objects.QueryResult {
	return printer_objects.QueryResult{
		"gcode_position":       object.printer.GcodeState.Position,
		"position":             object.printer.GcodeState.Position,
		"homing_origin":        [4]float64{0, 0, 0, 0},
		"speed":                object.printer.GcodeState.Feedrate / 60.0,
		"speed_factor":         float64(object.printer.GcodeState.SpeedFactor) / 100.0,
		"extrude_factor":       float64(object.printer.GcodeState.ExtrudeFactor) / 100.0,
		"absolute_coordinates": object.printer.GcodeState.IsAbsoluteCoordinate,
		"absolute_extrude":     object.printer.GcodeState.IsAbsoluteExtrude,
	}
}
