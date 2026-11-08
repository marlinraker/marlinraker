package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerFilesRootsResult []files.FileRoot

func ServerFilesRoots(*connections.Connection, *http.Request, Params) (any, error) {
	return files.FileRoots, nil
}
