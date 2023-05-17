package printer

import (
	"marlinraker/src/printer/parser"
	"marlinraker/src/printer_objects"
	"marlinraker/src/shared"
	"math"
	"strconv"
	"strings"
)

type GcodeState struct {
	Position             [4]float32
	IsAbsoluteCoordinate bool
	IsAbsoluteExtrude    bool
	SpeedFactor          int32
	ExtrudeFactor        int32
	Feedrate             float32
	HomedAxes            [3]bool
}

func (state *GcodeState) update(line string) error {

	switch {
	case parser.G0_G1_G92.MatchString(line):
		isG92 := parser.G92.MatchString(line)
		values, err := parser.ParseG0G1G92(line)
		if err != nil {
			return err
		}
		for i := 0; i < 3; i++ {
			axis := string("XYZ"[i])
			if value, exists := values[axis]; exists {
				if state.IsAbsoluteCoordinate || isG92 {
					state.Position[i] = value
				} else {
					state.Position[i] += value
				}
			}
		}
		if value, exists := values["E"]; exists {
			if state.IsAbsoluteExtrude || isG92 {
				state.Position[3] = value
			} else {
				state.Position[3] += value
			}
		}
		if value, exists := values["F"]; exists && !isG92 {
			state.Feedrate = value
		}
		if err := printer_objects.EmitObject("toolhead"); err != nil {
			return err
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

	case parser.G91.MatchString(line):
		state.IsAbsoluteCoordinate = false
		state.IsAbsoluteExtrude = false

	case parser.M82.MatchString(line):
		state.IsAbsoluteExtrude = true

	case parser.M83.MatchString(line):
		state.IsAbsoluteExtrude = false

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
		builder.WriteString("M220 S" + strconv.Itoa(int(restoreTo.SpeedFactor)) + "\n")
		state.SpeedFactor = restoreTo.SpeedFactor
	}

	if state.ExtrudeFactor != restoreTo.ExtrudeFactor {
		builder.WriteString("M221 S" + strconv.Itoa(int(restoreTo.ExtrudeFactor)) + "\n")
		state.ExtrudeFactor = restoreTo.ExtrudeFactor
	}

	if state.Feedrate != restoreTo.Feedrate {
		builder.WriteString("G0 F" + strconv.Itoa(int(restoreTo.Feedrate)) + "\n")
		state.Feedrate = restoreTo.Feedrate
	}

	coords := make([]string, 0)
	for i, to := range restoreTo.Position {
		from := state.Position[i]
		if math.Abs(float64(from-to)) > 1e-6 {
			var value float64
			if state.IsAbsoluteExtrude {
				value = float64(to)
			} else {
				value = float64(to - from)
			}
			axis := string("XYZE"[i])
			coord := axis + strconv.FormatFloat(value, 'f', 3, 32)
			coords = append(coords, coord)
		}
	}

	if len(coords) > 0 {
		builder.WriteString("G1 " + strings.Join(coords, " ") + "\n")
	}
	state.Position = restoreTo.Position

	<-context.QueueGcode(builder.String(), true)
}
