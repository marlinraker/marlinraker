package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/util"
	"net/http"
	"path/filepath"
	"strings"
)

type ServerFilesDeleteFileResult files.FileDeleteAction

func ServerFilesDeleteFile(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	path, exists := params.GetString("path")
	if !exists {
		return nil, util.NewError("path param is required", 400)
	}

	path = strings.TrimPrefix(path, "/")
	root := strings.Split(path, "/")[0]
	file := strings.TrimPrefix(filepath.Clean(path[len(root):]), "/")
	return files.DeleteFile(root, file)
}
