package macros

import (
	"errors"
	"github.com/samber/lo"
	"marlinraker/src/shared"
	"strconv"
	"strings"
)

type setHeaterTemperatureMacro struct{}

func (setHeaterTemperatureMacro) Description() string {
	return "Sets a heater temperature"
}

func (setHeaterTemperatureMacro) Execute(_ *MacroManager, context shared.ExecutorContext, _ []string, objects Objects, params Params) error {

	heater, err := params.RequireString("heater")
	if err != nil {
		return err
	}

	availableHeaters := objects["heaters"]["available_heaters"].([]string)
	if !lo.Contains(availableHeaters, heater) {
		return errors.New("cannot find heater \"" + heater + "\"")
	}

	target, err := params.RequireFloat64("target")
	if err != nil {
		return err
	}

	switch {
	case strings.HasPrefix(heater, "extruder"):
		idx := heater[8:]
		if idx == "" {
			idx = "0"
		}
		gcode := "M104 T" + idx + " S" + strconv.FormatFloat(target, 'f', 2, 64)
		<-context.QueueGcode(gcode, true)

	case heater == "heater_bed":
		gcode := "M140 S" + strconv.FormatFloat(target, 'f', 2, 64)
		<-context.QueueGcode(gcode, true)

	default:
		return errors.New("could not map heater name \"" + heater + "\"")
	}
	return nil
}
