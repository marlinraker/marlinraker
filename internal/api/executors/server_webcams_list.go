package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/webcams"
)

type ServerWebcamsListResult struct {
	Webcams []webcams.Webcam `json:"webcams"`
}

func ServerWebcamsList(*connections.Connection, *http.Request, Params) (any, error) {
	list, err := webcams.List()
	if err != nil {
		return nil, err
	}
	return ServerWebcamsListResult{Webcams: list}, nil
}
