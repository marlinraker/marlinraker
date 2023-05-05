package executors

import (
	"marlinraker/src/constants"
	"marlinraker/src/files"
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"net/http"
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
	return ServerInfoResult{
		KlippyConnected:           true,
		KlippyState:               string(marlinraker.State),
		Components:                []string{"server", "file_manager", "machine", "database", "data_store", "proc_stats", "history"},
		FailedComponents:          []string{},
		RegisteredDirectories:     files.GetRegisteredDirectories(),
		Warnings:                  []string{},
		WebsocketCount:            len(connections.GetConnections()),
		MissingKlippyRequirements: []string{},
		MoonrakerVersion:          constants.Version,
		ApiVersion:                constants.ApiVersion,
		ApiVersionString:          constants.ApiVersionString,
		Type:                      "marlinraker",
	}, nil
}
