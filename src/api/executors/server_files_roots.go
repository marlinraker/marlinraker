package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
)

type ServerFilesRootsResult []files.FileRoot

func ServerFilesRoots(*connections.Connection, Params) (any, error) {
	return files.FileRoots, nil
}
