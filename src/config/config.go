package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	resources "marlinraker"
	"marlinraker/src/files"
	"path/filepath"
	"regexp"
)

type Web struct {
	BindAddress string   `toml:"bind_address"`
	Port        int      `toml:"port"`
	CorsDomains []string `toml:"cors_domains"`
}

type Serial struct {
	Port                  string      `toml:"port"`
	BaudRate              interface{} `toml:"baud_rate"`
	MaxConnectionAttempts int         `toml:"max_connection_attempts"`
	ConnectionTimeout     int         `toml:"connection_timeout"`
}

type Misc struct {
	OctoprintCompat bool     `toml:"octoprint_compat"`
	ExtendedLogs    bool     `toml:"extended_logs"`
	ReportVelocity  bool     `toml:"report_velocity"`
	AllowedServices []string `toml:"allowed_services"`
}

type Heater struct {
	MinTemp int `toml:"min_temp"`
	MaxTemp int `toml:"max_temp"`
}

type Extruder struct {
	Heater
	MinExtrudeTemp   int     `toml:"min_extrude_temp"`
	FilamentDiameter float32 `toml:"filament_diameter"`
}

type HeaterBed struct {
	Heater
}

type Gcode struct {
	SendM73 bool `toml:"send_m73"`
}

type Printer struct {
	BedMesh     bool      `toml:"bed_mesh"`
	PrintVolume [3]int    `toml:"print_volume"`
	Extruder    Extruder  `toml:"extruder"`
	HeaterBed   HeaterBed `toml:"heater_bed"`
	Gcode       Gcode     `toml:"gcode"`
}

type Macro struct {
	RenameExisting string `toml:"rename_existing"`
	Gcode          string `toml:"gcode"`
}

type Config struct {
	Web     Web              `toml:"web"`
	Serial  Serial           `toml:"serial"`
	Misc    Misc             `toml:"misc"`
	Printer Printer          `toml:"printer"`
	Macros  map[string]Macro `toml:"macros"`
}

var includeRegex = regexp.MustCompile(`(?mi)^#<include +([\w\-. /\\]+?)>.*$`)

func CopyDefaults(targetDir string) error {
	configPath := filepath.Join(targetDir, "marlinraker.toml")

	if _, err := files.Fs.Stat(configPath); err != nil {
		err := afero.WriteFile(files.Fs, configPath, []byte(resources.ExampleConfig), 0755)
		if err != nil {
			return err
		}

		printerConfigPath := filepath.Join(targetDir, "printer.toml")
		if _, err := files.Fs.Stat(printerConfigPath); err != nil {
			err = afero.WriteFile(files.Fs, printerConfigPath, []byte(resources.ExamplePrinterConfig), 0755)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func LoadConfig(path string) (*Config, error) {
	contents, err := resolve(path, []string{})
	if err != nil {
		return nil, err
	}
	return parseConfig(contents)
}

func resolve(currentPath string, resolvedSoFar []string) (string, error) {

	if !filepath.IsAbs(currentPath) {
		currentPath, _ = filepath.Abs(currentPath)
	}

	contentBytes, err := afero.ReadFile(files.Fs, currentPath)
	if err != nil {
		return "", err
	}
	content := string(contentBytes)
	resolvedSoFar = append(resolvedSoFar, currentPath)

	matches := includeRegex.FindAllStringIndex(content, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		start, end := matches[i][0], matches[i][1]
		str := content[start:end]
		filename := includeRegex.FindStringSubmatch(str)[1]
		nextPath, err := filepath.Abs(filepath.Join(filepath.Dir(currentPath), filename))
		if err != nil {
			return "", err
		}

		if lo.Contains(resolvedSoFar, nextPath) {
			return "", errors.New("Error in " + currentPath + ": Cannot resolve cyclic dependency \"" + filename + "\"")
		}

		result, err := resolve(nextPath, resolvedSoFar)
		if err != nil {
			return result, err
		}

		content = content[0:start] + result + content[end:]
	}

	return content, nil
}

func DefaultConfig() *Config {
	return &Config{
		Web: Web{
			Port:        7125,
			CorsDomains: []string{},
		},
		Serial: Serial{
			Port:                  "auto",
			BaudRate:              "auto",
			MaxConnectionAttempts: 5,
			ConnectionTimeout:     5000,
		},
		Misc: Misc{
			OctoprintCompat: true,
			ExtendedLogs:    false,
			ReportVelocity:  false,
			AllowedServices: []string{"marlinraker", "crowsnest", "MoonCord", "moonraker-telegram-bot", "KlipperScreen", "sonar", "webcamd"},
		},
		Printer: Printer{
			BedMesh:     false,
			PrintVolume: [3]int{220, 220, 240},
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
				SendM73: true,
			},
		},
		Macros: map[string]Macro{},
	}
}

func parseConfig(contents string) (*Config, error) {
	config := DefaultConfig()
	_, err := toml.Decode(contents, config)
	return config, err
}
