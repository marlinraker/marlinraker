package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerFilesMetadataResult *files.Metadata

func ServerFilesMetadata(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	fileName, err := params.RequirePath("filename")
	if err != nil {
		return nil, err
	}

	return files.LoadOrScanMetadata(fileName)
}
