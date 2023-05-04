package executors

import (
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/util"
	"net/http"
)

type ServerConnectionIdentifyResult struct {
	ConnectionId int `json:"connection_id"`
}

func ServerConnectionIdentify(connection *connections.Connection, _ *http.Request, params Params) (any, error) {

	if connection.Identified {
		return nil, util.NewError("connection already identified", 400)
	}

	clientName, exists := params["client_name"].(string)
	if !exists {
		return nil, util.NewError("client_name param is required", 400)
	}

	version, exists := params["version"].(string)
	if !exists {
		return nil, util.NewError("version param is required", 400)
	}

	clientType, exists := params["type"].(string)
	if !exists {
		return nil, util.NewError("type param is required", 400)
	}

	url, exists := params["url"].(string)
	if !exists {
		return nil, util.NewError("url param is required", 400)
	}

	connection.ClientName, connection.Version, connection.ClientType, connection.Url =
		clientName, version, clientType, url

	return ServerConnectionIdentifyResult{connection.Id}, nil
}
