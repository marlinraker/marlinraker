package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
)

type ServerFilesGetDirectoryResult files.DirectoryInfo

func ServerFilesGetDirectory(_ *connections.Connection, params Params) (any, error) {
	path, exists := params["path"].(string)
	if !exists {
		path = "gcodes"
	}

	return files.GetDirInfo(path)
}
