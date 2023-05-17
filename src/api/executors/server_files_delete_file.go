package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type ServerFilesDeleteFileResult files.FileDeleteAction

func ServerFilesDeleteFile(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	path, err := params.RequirePath("path")
	if err != nil {
		return nil, err
	}
	return files.DeleteFile(path)
}
