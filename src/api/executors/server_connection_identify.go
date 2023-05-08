package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

type ServerConnectionIdentifyResult struct {
	ConnectionId int `json:"connection_id"`
}

func ServerConnectionIdentify(connection *connections.Connection, _ *http.Request, params Params) (any, error) {

	if connection.Identified {
		return nil, util.NewError("connection already identified", 400)
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

	connection.ClientName, connection.Version, connection.ClientType, connection.Url =
		clientName, version, clientType, url

	return ServerConnectionIdentifyResult{connection.Id}, nil
}
