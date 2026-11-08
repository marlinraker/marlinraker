package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerFilesGetDirectoryResult files.DirectoryInfo

func ServerFilesGetDirectory(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	path, err := params.RequirePath("path")
	if err != nil {
		return nil, err
	}

	extended, _ := params.GetBool("extended")

	return files.GetDirInfo(path, extended)
}
