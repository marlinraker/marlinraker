package marlinraker

import "marlinraker/src/printer_objects"

type configFileObject struct{}

func (configfileObject configFileObject) Query() printer_objects.QueryResult {
	return printer_objects.QueryResult{
		"settings":                  FakeKlipperConfig,
		"config":                    FakeKlipperConfig,
		"save_config_pending":       false,
		"save_config_pending_items": []string{},
		"warnings":                  []string{},
	}
}
