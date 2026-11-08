package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerFilesMoveResult files.MoveAction

func ServerFilesMove(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	source, err := params.RequirePath("source")
	if err != nil {
		return nil, err
	}

	dest, err := params.RequirePath("dest")
	if err != nil {
		return nil, err
	}

	return files.Move(source, dest)
}
