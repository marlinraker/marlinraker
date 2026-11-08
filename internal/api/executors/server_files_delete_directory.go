package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerFilesDeleteDirectoryResult files.DirectoryAction

func ServerFilesDeleteDirectory(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	path, err := params.RequirePath("path")
	if err != nil {
		return nil, err
	}

	force, _ := params.GetBool("force")

	return files.DeleteDir(path, force)
}
