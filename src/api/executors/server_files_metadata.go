package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type ServerFilesMetadataResult *files.Metadata

func ServerFilesMetadata(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	fileName, err := params.RequireString("filename")
	if err != nil {
		return nil, err
	}

	return files.LoadOrScanMetadata(fileName)
}
