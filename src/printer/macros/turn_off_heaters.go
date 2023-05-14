package macros

import (
	"marlinraker/src/shared"
	"strings"
)

type turnOffHeatersMacro struct{}

func (turnOffHeatersMacro) Description() string {
	return "Turn off all heaters"
}

func (turnOffHeatersMacro) Execute(_ *MacroManager, context shared.ExecutorContext, _ []string, objects Objects, _ Params) error {

	availableHeaters := objects["heaters"]["available_heaters"].([]string)
	for _, heater := range availableHeaters {
		switch {
		case strings.HasPrefix(heater, "extruder"):
			idx := heater[8:]
			if idx == "" {
				idx = "0"
			}
			<-context.QueueGcode("M104 T"+idx+" S0", true)

		case heater == "heater_bed":
			<-context.QueueGcode("M140 S0", true)
		}
	}

	return nil
}
