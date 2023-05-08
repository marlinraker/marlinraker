package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"net/http"
	"path/filepath"
	"strings"
)

type ServerFilesDeleteFileResult files.FileDeleteAction

func ServerFilesDeleteFile(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	path, err := params.RequireString("path")
	if err != nil {
		return nil, err
	}

	path = strings.TrimPrefix(path, "/")
	root := strings.Split(path, "/")[0]
	file := strings.TrimPrefix(filepath.Clean(path[len(root):]), "/")
	return files.DeleteFile(root, file)
}
