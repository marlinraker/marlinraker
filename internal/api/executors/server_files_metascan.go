package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerFilesMetascanResult *files.Metadata

func ServerFilesMetascan(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	fileName, err := params.RequirePath("filename")
	if err != nil {
		return nil, err
	}

	_ = files.RemoveMetadata(fileName)

	return files.LoadOrScanMetadata(fileName)
}
