package api

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/api/executors"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/printer_objects"
	"marlinraker/src/util"
	"net/http"
)

type Rpc struct {
	JsonRpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
}

type RpcRequest struct {
	Rpc
	Method string           `json:"method"`
	Params executors.Params `json:"params"`
}

type RpcResultResponse struct {
	Rpc
	Result any `json:"result"`
}

type RpcErrorResponse struct {
	Rpc
	Error Error `json:"error"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleSocket(writer http.ResponseWriter, request *http.Request) error {
	socket, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		return err
	}

	connection := connections.RegisterConnection(socket)

	for {
		_, message, err := socket.ReadMessage()
		if err != nil {
			break
		}
		log.Debugln("recv: " + string(message))

		var request RpcRequest
		err = json.Unmarshal(message, &request)
		if err != nil {
			util.LogError(err)
			continue
		}

		executor := socketExecutors[request.Method]
		if executor == nil {
			log.Errorln("No executor found for " + request.Method)
			err = connection.WriteJson(&RpcErrorResponse{
				Error: Error{404, "Not Found"},
				Rpc:   request.Rpc,
			})
			if err != nil {
				util.LogError(err)
			}
			continue
		}

		result, err := executor(connection, nil, request.Params)
		if err != nil {
			log.Errorln("Error while executing "+request.Method+":", err)
			code := 500
			if executorError, isExecutorError := err.(*util.ExecutorError); isExecutorError {
				code = executorError.Code
			}
			err = connection.WriteJson(&RpcErrorResponse{
				Error: Error{code, err.Error()},
				Rpc:   request.Rpc,
			})
			if err != nil {
				util.LogError(err)
			}
			continue
		}

		err = connection.WriteJson(&RpcResultResponse{request.Rpc, result})
		if err != nil {
			util.LogError(err)
		}
	}

	connections.UnregisterConnection(connection)
	printer_objects.Unsubscribe(connection)
	return nil
}
