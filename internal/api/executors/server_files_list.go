package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/util"
)

type ServerFilesResult = []files.File

func ServerFilesList(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	root, exists := params.GetString("root")
	if !exists {
		root = "gcodes"
	}
	filesList, err := files.ListFiles(root)
	if err != nil {
		return nil, util.NewErrorf(400, "%v", err)
	}
	return filesList, nil
}
