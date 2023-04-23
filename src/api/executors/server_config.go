package executors

import (
	"marlinraker-go/src/config"
	"marlinraker-go/src/marlinraker"
	"marlinraker-go/src/marlinraker/connections"
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

func ServerConfig(*connections.Connection, Params) (any, error) {
	return ServerConfigResult{
		Config: marlinraker.Config,
		Orig:   marlinraker.Config,
		Files:  []ConfigFile{},
	}, nil
}
