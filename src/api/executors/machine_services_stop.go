package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/service"
	"net/http"
)

func MachineServicesStop(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	svc, err := params.RequireString("service")
	if err != nil {
		return nil, err
	}

	if err := service.PerformAction(svc, service.Stop); err != nil {
		return nil, err
	}
	return "ok", nil
}
