package macros

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/marlinraker/marlinraker/internal/shared"
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
		return fmt.Errorf("cannot find heater %q", heater)
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
		gcode := fmt.Sprintf("M104 T%s S%.2f", idx, target)
		<-context.QueueGcode(gcode, true)

	case heater == "heater_bed":
		gcode := fmt.Sprintf("M140 S%.2f", target)
		<-context.QueueGcode(gcode, true)

	default:
		return fmt.Errorf("could not map heater name %q", heater)
	}
	return nil
}
