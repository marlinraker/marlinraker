package executors

import (
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/util"
	"net/http"
)

type ServerFilesMoveResult files.MoveAction

func ServerFilesMove(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	source, exists := params["source"].(string)
	if !exists {
		return nil, util.NewError("source param is required", 400)
	}

	dest, exists := params["dest"].(string)
	if !exists {
		return nil, util.NewError("dest param is required", 400)
	}

	return files.Move(source, dest)
}
