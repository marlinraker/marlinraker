package printer

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/marlinraker/marlinraker/internal/printer/parser"
	"github.com/marlinraker/marlinraker/internal/printer_objects"
	"github.com/marlinraker/marlinraker/internal/shared"
)

type GcodeState struct {
	Position             [4]float64
	IsAbsoluteCoordinate bool
	IsAbsoluteExtrude    bool
	SpeedFactor          int
	ExtrudeFactor        int
	Feedrate             float64
	HomedAxes            [3]bool
	EOffset              float64
	Velocity             float64
	EVelocity            float64
}

func (state *GcodeState) ExtrudedFilament() float64 {
	return state.EOffset + state.Position[3]
}

func (state *GcodeState) update(line string) error {

	switch {
	case parser.G92.MatchString(line):
		values, err := parser.ParseG0G1G92(line)
		if err != nil {
			return fmt.Errorf("failed to parse extruder offset: %w", err)
		}
		if value, exists := values["E"]; exists {
			state.EOffset += state.Position[3] - value
		}

	case parser.G28.MatchString(line):
		homedAxes := parser.ParseG28(line)
		for i := 0; i < 3; i++ {
			if homedAxes[i] {
				state.HomedAxes[i] = true
			}
		}
		if err := printer_objects.EmitObject("toolhead"); err != nil {
			return err
		}

	case parser.M18_M84_M410.MatchString(line):
		for i := 0; i < 3; i++ {
			state.HomedAxes[i] = false
		}
		if err := printer_objects.EmitObject("toolhead"); err != nil {
			return err
		}

	case parser.G90.MatchString(line):
		state.IsAbsoluteCoordinate = true
		state.IsAbsoluteExtrude = true
		if err := printer_objects.EmitObject("gcode_move"); err != nil {
			return err
		}

	case parser.G91.MatchString(line):
		state.IsAbsoluteCoordinate = false
		state.IsAbsoluteExtrude = false
		if err := printer_objects.EmitObject("gcode_move"); err != nil {
			return err
		}

	case parser.M82.MatchString(line):
		state.IsAbsoluteExtrude = true
		if err := printer_objects.EmitObject("gcode_move"); err != nil {
			return err
		}

	case parser.M83.MatchString(line):
		state.IsAbsoluteExtrude = false
		if err := printer_objects.EmitObject("gcode_move"); err != nil {
			return err
		}

	case parser.M220_M221.MatchString(line):
		factor, err := parser.ParseM220M221(line)
		if err != nil {
			return err
		}
		if parser.M220.MatchString(line) {
			state.SpeedFactor = factor
		} else {
			state.ExtrudeFactor = factor
		}
		if err := printer_objects.EmitObject("gcode_move"); err != nil {
			return err
		}
	}

	return nil
}

func (state *GcodeState) restore(context shared.ExecutorContext, restoreTo GcodeState) {

	var builder strings.Builder
	if restoreTo.IsAbsoluteCoordinate {
		builder.WriteString("G90\n")
		if !restoreTo.IsAbsoluteExtrude {
			builder.WriteString("M83\n")
		}
	} else {
		builder.WriteString("G91\n")
		if restoreTo.IsAbsoluteExtrude {
			builder.WriteString("M82\n")
		}
	}
	state.IsAbsoluteCoordinate, state.IsAbsoluteExtrude =
		restoreTo.IsAbsoluteCoordinate, restoreTo.IsAbsoluteExtrude

	if state.SpeedFactor != restoreTo.SpeedFactor {
		fmt.Fprintf(&builder, "M220 S%d\n", restoreTo.SpeedFactor)
		state.SpeedFactor = restoreTo.SpeedFactor
	}

	if state.ExtrudeFactor != restoreTo.ExtrudeFactor {
		fmt.Fprintf(&builder, "M221 S%d\n", restoreTo.ExtrudeFactor)
		state.ExtrudeFactor = restoreTo.ExtrudeFactor
	}

	if state.Feedrate != restoreTo.Feedrate {
		fmt.Fprintf(&builder, "G0 F%d\n", int(restoreTo.Feedrate))
		state.Feedrate = restoreTo.Feedrate
	}

	coords := make([]string, 0)
	for i, to := range restoreTo.Position {
		from := state.Position[i]
		if math.Abs(from-to) > 1e-6 {
			var value float64
			if state.IsAbsoluteExtrude {
				value = to
			} else {
				value = to - from
			}
			axis := string("XYZE"[i])
			coord := axis + strconv.FormatFloat(value, 'f', 3, 32)
			coords = append(coords, coord)
		}
	}

	if len(coords) > 0 {
		fmt.Fprintf(&builder, "G1 %s\n", strings.Join(coords, " "))
	}
	state.Position = restoreTo.Position

	<-context.QueueGcode(builder.String(), true)
}
