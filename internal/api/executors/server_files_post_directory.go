package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerFilesPostDirectoryResult files.DirectoryAction

func ServerFilesPostDirectory(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	path, err := params.RequirePath("path")
	if err != nil {
		return nil, err
	}

	return files.CreateDir(path)
}
