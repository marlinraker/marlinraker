package macros

import (
	"strings"
)

type turnOffHeatersMacro struct{}

func (turnOffHeatersMacro) Description() string {
	return "Turn off all heaters"
}

func (turnOffHeatersMacro) Execute(manager *MacroManager, _ []string, objects Objects, _ Params) error {

	availableHeaters := objects["heaters"]["available_heaters"].([]string)
	for _, heater := range availableHeaters {
		switch {
		case strings.HasPrefix(heater, "extruder"):
			idx := heater[8:]
			if idx == "" {
				idx = "0"
			}
			<-manager.printer.QueueGcode("M104 T"+idx+" S0", false, true)

		case heater == "heater_bed":
			<-manager.printer.QueueGcode("M140 S0", false, true)
		}
	}

	return nil
}
