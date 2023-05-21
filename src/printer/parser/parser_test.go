package parser

import (
	"github.com/spf13/afero"
	"gotest.tools/assert"
	"marlinraker/src/files"
	"strconv"
	"strings"
	"testing"
)

func readContent(t *testing.T, name string) string {
	bytes, err := afero.ReadFile(files.Fs, name)
	assert.NilError(t, err)
	return string(bytes)
}

func TestParseM115(t *testing.T) {
	response := readContent(t, "testdata/m115")
	info, capabilities, err := ParseM115(response)
	assert.NilError(t, err)

	assert.DeepEqual(t, info, PrinterInfo{
		MachineType: "Prusa-mini", FirmwareName: "Prusa-Firmware-Buddy 4.4.0-BETA2+4114 (Github)",
	})

	assert.DeepEqual(t, capabilities, map[string]bool{
		"SERIAL_XON_XOFF":       false,
		"BINARY_FILE_TRANSFER":  false,
		"EEPROM":                false,
		"VOLUMETRIC":            true,
		"AUTOREPORT_TEMP":       true,
		"PROGRESS":              false,
		"PRINT_JOB":             true,
		"AUTOLEVEL":             true,
		"Z_PROBE":               true,
		"LEVELING_DATA":         true,
		"BUILD_PERCENT":         false,
		"SOFTWARE_POWER":        false,
		"TOGGLE_LIGHTS":         false,
		"CASE_LIGHT_BRIGHTNESS": false,
		"EMERGENCY_PARSER":      false,
		"PROMPT_SUPPORT":        true,
		"AUTOREPORT_SD_STATUS":  false,
		"THERMAL_PROTECTION":    true,
		"MOTION_MODES":          false,
		"CHAMBER_TEMPERATURE":   false,
	})
}

func TestParseM503(t *testing.T) {
	response := readContent(t, "testdata/m503")
	limits, err := ParseM503(response)
	assert.NilError(t, err)

	assert.DeepEqual(t, limits, PrinterLimits{
		MaxAccel:    [3]float32{1250.00, 1250.00, 400.00},
		MaxFeedrate: [3]float32{180.00, 180.00, 12.00},
	})
}

func TestParseM105(t *testing.T) {
	responseLines := readContent(t, "testdata/m105")

	expected := []map[string]any{
		{ // 1
			"extruder":  Heater{Sensor{17.54}, 0, 0},
			"extruder1": Heater{Sensor{20.7}, 0, 0},
		},
		{ // 2
			"extruder":                   Heater{Sensor{229}, 230, 55.0 / 127.0},
			"heater_bed":                 Heater{Sensor{84.96}, 85, 58.0 / 127.0},
			"temperature_sensor ambient": Sensor{-30},
		},
		{ // 3
			"extruder":   Sensor{201},
			"heater_bed": Sensor{117},
		},
		{ // 4
			"extruder":   Heater{Sensor{201}, 202, 0},
			"heater_bed": Heater{Sensor{117}, 120, 0},
		},
		{ // 5
			"extruder":               Heater{Sensor{201}, 202, 0},
			"heater_bed":             Heater{Sensor{117}, 120, 0},
			"heater_generic chamber": Heater{Sensor{49.3}, 50, 0},
		},
		{ // 6
			"extruder":               Heater{Sensor{110}, 110, 0},
			"extruder1":              Heater{Sensor{23}, 0, 0},
			"heater_bed":             Heater{Sensor{117}, 120, 0},
			"heater_generic chamber": Heater{Sensor{49.3}, 50, 0},
		},
		{ // 7
			"extruder":   Heater{Sensor{110}, 110, 0},
			"extruder1":  Heater{Sensor{23}, 0, 0},
			"heater_bed": Heater{Sensor{117}, 120, 0},
		},
		{ // 8
			"extruder":                   Heater{Sensor{20.2}, 0, 0},
			"heater_bed":                 Heater{Sensor{19.1}, 0, 0},
			"temperature_sensor pinda":   Sensor{19.8},
			"temperature_sensor ambient": Sensor{26.4},
		},
	}

	for i, response := range strings.Split(strings.TrimSpace(responseLines), "\n") {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			temperatures, err := ParseM105(response)
			if err != nil {
				t.Fatal(err)
			}
			assert.DeepEqual(t, temperatures, expected[i])
		})
	}
}

func TestParseM220M221(t *testing.T) {
	s, err := ParseM220M221("M220 S80")
	assert.NilError(t, err)
	assert.Equal(t, s, 80)

	_, err = ParseM220M221("M220")
	assert.Error(t, err, "missing S parameter")
}

func TestParseG0G1G92(t *testing.T) {
	coords, err := ParseG0G1G92("G1 X109.73 Y119.168 E.29192")
	assert.NilError(t, err)
	assert.DeepEqual(t, coords, map[string]float64{
		"X": 109.73,
		"Y": 119.168,
		"E": 0.29192,
	})
}

func TestParseM114(t *testing.T) {
	responseLines := readContent(t, "testdata/m114")
	expected := [][4]float64{
		{180.4, -3., 0., 0.},
		{0., 0., 0., 0.},
	}

	for i, line := range strings.Split(strings.TrimSpace(responseLines), "\n") {
		actual, err := ParseM114(line)
		assert.NilError(t, err)
		assert.DeepEqual(t, expected[i], actual)
	}
}
