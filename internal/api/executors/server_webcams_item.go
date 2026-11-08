package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/webcams"
)

type ServerWebcamsItemResult struct {
	Webcam webcams.Webcam `json:"webcam"`
}

func ServerWebcamsGetItem(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	uid, _ := params.GetString("uid")
	name, _ := params.GetString("name")
	webcam, err := webcams.GetItem(uid, name)
	if err != nil {
		return nil, err
	}
	return ServerWebcamsItemResult{Webcam: webcam}, nil
}

func ServerWebcamsPostItem(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	webcam, err := webcams.PostItem(params)
	if err != nil {
		return nil, err
	}
	return ServerWebcamsItemResult{Webcam: webcam}, nil
}

func ServerWebcamsDeleteItem(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	uid, _ := params.GetString("uid")
	name, _ := params.GetString("name")
	webcam, err := webcams.DeleteItem(uid, name)
	if err != nil {
		return nil, err
	}
	return ServerWebcamsItemResult{Webcam: webcam}, nil
}
