package marlinraker

import "marlinraker-go/src/printer_objects"

type webhooksObject struct{}

func (webhooksObject) Query() printer_objects.QueryResult {
	return printer_objects.QueryResult{
		"state":         State,
		"state_message": StateMessage,
	}
}
