package marlinraker

import "marlinraker/src/printer_objects"

type webhooksObject struct{}

func (webhooksObject) Query() (printer_objects.QueryResult, error) {
	state, message := GetState()
	return printer_objects.QueryResult{
		"state":         state,
		"state_message": message,
	}, nil
}
