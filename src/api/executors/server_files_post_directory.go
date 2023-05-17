package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type ServerFilesPostDirectoryResult files.DirectoryAction

func ServerFilesPostDirectory(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	path, err := params.RequirePath("path")
	if err != nil {
		return nil, err
	}

	return files.CreateDir(path)
}
