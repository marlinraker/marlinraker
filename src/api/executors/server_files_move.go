package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type ServerFilesMoveResult files.MoveAction

func ServerFilesMove(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	source, err := params.RequireString("source")
	if err != nil {
		return nil, err
	}

	dest, err := params.RequireString("dest")
	if err != nil {
		return nil, err
	}

	return files.Move(source, dest)
}
