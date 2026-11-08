package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/service"
)

func MachineServicesStart(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	svc, err := params.RequireString("service")
	if err != nil {
		return nil, err
	}

	if err := service.PerformAction(svc, service.Start); err != nil {
		return nil, err
	}
	return "ok", nil
}
