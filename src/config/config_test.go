package config

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"gotest.tools/assert"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("testdata/test_config.toml")
	if err != nil {
		t.Fatal(err)
		return
	}

	assert.DeepEqual(t, config, &Config{
		Web: Web{
			BindAddress: "1.2.3.4",
			Port:        123,
			CorsDomains: []string{"domain"},
		},
		Serial: Serial{
			Port:                  "auto",
			BaudRate:              int64(115200),
			MaxConnectionAttempts: 5,
			ConnectionTimeout:     5000,
		},
		Misc: Misc{
			OctoprintCompat: true,
			ExtendedLogs:    false,
			AllowedServices: []string{"service1", "service2"},
		},
		Printer: Printer{
			BedMesh:     false,
			AxisMinimum: [3]int{0, 0, 0},
			AxisMaximum: [3]int{220, 220, 240},
			Extruder: Extruder{
				Heater: Heater{
					MinTemp: 0,
					MaxTemp: 250,
				},
				MinExtrudeTemp:   180,
				FilamentDiameter: 1.75,
			},
			HeaterBed: HeaterBed{
				Heater: Heater{
					MinTemp: 0,
					MaxTemp: 100,
				},
			},
			Gcode: Gcode{
				SendM73:        true,
				ReportVelocity: true,
			},
		},
		Macros: map[string]Macro{
			"start_print": {
				RenameExisting: "start_base",
				Gcode:          "multiline\nmacro\n",
			},
			"test": {
				Gcode: "another test macro",
			},
		},
	})
}

func TestResolve(t *testing.T) {
	t.Run("Import resolving", func(t *testing.T) {
		result, err := resolve("testdata/valid/config1.txt", []string{})
		if err != nil {
			t.Fatal(err)
			return
		}

		assert.Equal(t, result, "Hello\nworld\n:)")
	})

	t.Run("Cyclic dependency detection", func(t *testing.T) {

		hook := test.Hook{}
		logrus.AddHook(&hook)

		_, err := resolve("testdata/invalid/config1.txt", []string{})
		assert.NilError(t, err)

		message := hook.LastEntry().Message
		assert.Assert(t, strings.HasPrefix(message, "Cannot resolve cyclic dependency"))
	})
}
