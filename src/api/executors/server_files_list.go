package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
)

type ServerFilesResult = []files.File

func ServerFilesList(_ *connections.Connection, params Params) (any, error) {
	root, exists := params["root"].(string)
	if !exists {
		root = "gcodes"
	}
	filesList, err := files.ListFiles(root)
	if err != nil {
		return nil, NewError(err.Error(), 400)
	}
	return filesList, nil
}
