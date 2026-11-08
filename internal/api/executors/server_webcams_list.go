package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerWebcamsListResult struct {
	// TODO
	Webcams []any `json:"webcams"`
}

func ServerWebcamsList(*connections.Connection, *http.Request, Params) (any, error) {
	return ServerWebcamsListResult{
		Webcams: []any{},
	}, nil
}
