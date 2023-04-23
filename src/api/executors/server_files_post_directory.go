package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
)

type ServerFilesPostDirectoryResult files.DirectoryAction

func ServerFilesPostDirectory(_ *connections.Connection, params Params) (any, error) {
	path, exists := params["path"].(string)
	if !exists {
		return nil, NewError("path param is required", 400)
	}

	return files.CreateDir(path)
}
