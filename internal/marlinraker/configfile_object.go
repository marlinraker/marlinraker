package marlinraker

import "github.com/marlinraker/marlinraker/internal/printer_objects"

type configFileObject struct{}

func (configfileObject configFileObject) Query() (printer_objects.QueryResult, error) {
	return printer_objects.QueryResult{
		"settings":                  KlipperSettings,
		"config":                    KlipperConfig,
		"save_config_pending":       false,
		"save_config_pending_items": []string{},
		"warnings":                  []string{},
	}, nil
}
