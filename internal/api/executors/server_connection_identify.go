package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/util"
)

type ServerConnectionIdentifyResult struct {
	ConnectionId int `json:"connection_id"`
}

func ServerConnectionIdentify(connection *connections.Connection, _ *http.Request, params Params) (any, error) {

	if connection == nil {
		return nil, util.NewError(400, "server.connection.identify is only available via websocket")
	}

	clientName, err := params.RequireString("client_name")
	if err != nil {
		return nil, err
	}

	version, err := params.RequireString("version")
	if err != nil {
		return nil, err
	}

	clientType, err := params.RequireString("type")
	if err != nil {
		return nil, err
	}

	url, err := params.RequireString("url")
	if err != nil {
		return nil, err
	}

	info := connections.ClientInfo{ClientName: clientName, Version: version, ClientType: clientType, Url: url}
	if !connection.Identify(info) {
		return nil, util.NewError(400, "connection already identified")
	}

	return ServerConnectionIdentifyResult{connection.Id}, nil
}
