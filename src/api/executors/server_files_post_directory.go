package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

type ServerFilesPostDirectoryResult files.DirectoryAction

func ServerFilesPostDirectory(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	path, exists := params.GetString("path")
	if !exists {
		return nil, util.NewError("path param is required", 400)
	}

	return files.CreateDir(path)
}
