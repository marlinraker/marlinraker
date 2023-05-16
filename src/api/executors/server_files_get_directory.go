package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type ServerFilesGetDirectoryResult files.DirectoryInfo

func ServerFilesGetDirectory(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	path, err := params.RequireString("path")
	if err != nil {
		return nil, err
	}

	extended, _ := params.GetBool("extended")

	return files.GetDirInfo(path, extended)
}
