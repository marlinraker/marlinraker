package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/util"
	"net/http"
)

type ServerFilesDeleteDirectoryResult files.DirectoryAction

func ServerFilesDeleteDirectory(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	path, exists := params.GetString("path")
	if !exists {
		return nil, util.NewError("path param is required", 400)
	}

	force, _ := params.GetBool("force")

	return files.DeleteDir(path, force)
}
