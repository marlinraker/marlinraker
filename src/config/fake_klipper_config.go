package config

import (
	"fmt"
	"strings"
)

func GenerateFakeKlipperConfig(config *Config) map[string]any {
	fakeConfig := map[string]any{
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
		fakeConfig["bed_mesh"] = map[string]any{
			"mesh_min": "0, 0",
			"mesh_max": fmt.Sprintf("%d, %d", config.Printer.PrintVolume[0], config.Printer.PrintVolume[1]),
		}
	}

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
		fakeConfig["gcode_macro "+strings.ToUpper(name)] = macroJson
	}

	return fakeConfig
}
