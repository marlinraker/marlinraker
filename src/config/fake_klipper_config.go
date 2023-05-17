package config

import (
	"fmt"
	"github.com/samber/lo"
	"strconv"
	"strings"
)

func GenerateFakeKlipperConfig(config *Config) (map[string]any, map[string]any) {

	settings := map[string]any{
		"extruder": map[string]any{
			"min_temp":          config.Printer.Extruder.MinTemp,
			"max_temp":          config.Printer.Extruder.MaxTemp,
			"min_extrude_temp":  config.Printer.Extruder.MinExtrudeTemp,
			"filament_diameter": config.Printer.Extruder.FilamentDiameter,
		},
		"heater_bed": map[string]any{
			"min_temp": config.Printer.HeaterBed.MinTemp,
			"max_temp": config.Printer.HeaterBed.MaxTemp,
		},
		"printer": map[string]any{
			"kinematics":     "cartesian",
			"max_velocity":   0,
			"max_accel":      0,
			"max_z_velocity": 0,
			"max_z_accel":    0,
		},
		"virtual_sdcard": map[string]any{},
		"pause_resume":   map[string]any{},
		"display_status": map[string]any{},
	}

	if config.Printer.BedMesh {
		settings["bed_mesh"] = map[string]any{
			"mesh_min": []int{0, 0},
			"mesh_max": config.Printer.PrintVolume[:2],
		}
	}

	configuration := stringifySettings(settings)

	for name, macro := range config.Macros {
		macroJson := map[string]any{
			"gcode": macro.Gcode,
		}
		if macro.RenameExisting != "" {
			macroJson["rename_existing"] = strings.ToUpper(macro.RenameExisting)
		}
		for name, value := range macro.Variables {
			macroJson["variable_"+name] = value
		}
		settings["gcode_macro "+strings.ToLower(name)] = macroJson
		configuration["gcode_macro "+strings.ToUpper(name)] = macroJson
	}

	return settings, configuration
}

func stringifySettings(settings map[string]any) map[string]any {
	configuration := make(map[string]any)
	for key, value := range settings {
		switch value := value.(type) {
		case map[string]any:
			configuration[key] = stringifySettings(value)
		case []int:
			configuration[key] = strings.Join(lo.Map(value, func(i int, _ int) string {
				return strconv.Itoa(i)
			}), ", ")
		default:
			configuration[key] = fmt.Sprintf("%v", value)
		}
	}
	return configuration
}
