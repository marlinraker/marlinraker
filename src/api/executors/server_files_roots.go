package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
	"net/http"
)

type ServerFilesRootsResult []files.FileRoot

func ServerFilesRoots(*connections.Connection, *http.Request, Params) (any, error) {
	return files.FileRoots, nil
}
