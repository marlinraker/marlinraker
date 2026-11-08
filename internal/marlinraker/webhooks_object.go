package marlinraker

import "github.com/marlinraker/marlinraker/internal/printer_objects"

type webhooksObject struct{}

func (webhooksObject) Query() (printer_objects.QueryResult, error) {
	state, message := GetState()
	return printer_objects.QueryResult{
		"state":         state,
		"state_message": message,
	}, nil
}
