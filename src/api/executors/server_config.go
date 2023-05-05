package executors

import (
	"marlinraker/src/config"
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type ConfigFile struct {
	Filename string   `json:"filename"`
	Sections []string `json:"sections"`
}

type ServerConfigResult struct {
	Config *config.Config `json:"config"`
	Orig   *config.Config `json:"orig"`
	Files  []ConfigFile   `json:"files"`
}

func ServerConfig(*connections.Connection, *http.Request, Params) (any, error) {
	return ServerConfigResult{
		Config: marlinraker.Config,
		Orig:   marlinraker.Config,
		Files:  []ConfigFile{},
	}, nil
}
