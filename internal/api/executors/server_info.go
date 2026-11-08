package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/constants"
	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerInfoResult struct {
	KlippyConnected           bool     `json:"klippy_connected"`
	KlippyState               string   `json:"klippy_state"`
	Components                []string `json:"components"`
	FailedComponents          []string `json:"failed_components"`
	RegisteredDirectories     []string `json:"registered_directories"`
	Warnings                  []string `json:"warnings"`
	WebsocketCount            int      `json:"websocket_count"`
	MissingKlippyRequirements []string `json:"missing_klippy_requirements"`
	MoonrakerVersion          string   `json:"moonraker_version"`
	ApiVersion                [3]int   `json:"api_version"`
	ApiVersionString          string   `json:"api_version_string"`
	Type                      string   `json:"type"`
}

func ServerInfo(*connections.Connection, *http.Request, Params) (any, error) {
	state, _ := marlinraker.GetState()
	return ServerInfoResult{
		KlippyConnected:           true,
		KlippyState:               string(state),
		Components:                []string{"server", "file_manager", "machine", "database", "data_store", "proc_stats", "history"},
		FailedComponents:          []string{},
		RegisteredDirectories:     files.GetRegisteredDirectories(),
		Warnings:                  []string{},
		WebsocketCount:            len(connections.GetConnections()),
		MissingKlippyRequirements: []string{},
		MoonrakerVersion:          constants.MoonrakerVersion(),
		ApiVersion:                constants.ApiVersion,
		ApiVersionString:          constants.ApiVersionString,
		Type:                      "marlinraker",
	}, nil
}
