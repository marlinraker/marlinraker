package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
	"net/http"
)

type ServerFilesGetDirectoryResult files.DirectoryInfo

func ServerFilesGetDirectory(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	path, exists := params.GetString("path")
	if !exists {
		path = "gcodes"
	}

	return files.GetDirInfo(path)
}
