package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
)

type ServerFilesDeleteDirectoryResult files.DirectoryAction

func ServerFilesDeleteDirectory(_ *connections.Connection, params Params) (any, error) {

	path, exists := params["path"].(string)
	if !exists {
		return nil, NewError("path param is required", 400)
	}

	forceParam, exists := params["force"]
	force := exists && (forceParam == true || forceParam == "true")

	return files.DeleteDir(path, force)
}
