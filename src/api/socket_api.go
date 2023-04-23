package api

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"marlinraker-go/src/api/executors"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/printer_objects"
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

func handleSocket(writer http.ResponseWriter, request *http.Request) {
	socket, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Error(err)
		return
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
			log.Error(err)
			continue
		}

		executor := socketExecutors[request.Method]
		if executor == nil {
			log.Errorln("No executor found for " + request.Method)
			_ = connection.WriteJson(&RpcErrorResponse{
				Error: Error{404, "Not Found"},
				Rpc:   request.Rpc,
			})
			continue
		}

		result, err := executor(connection, request.Params)
		if err != nil {
			log.Error(err)
			code := 500
			if executorError, isExecutorError := err.(*executors.ExecutorError); isExecutorError {
				code = executorError.Code
			}
			_ = connection.WriteJson(&RpcErrorResponse{
				Error: Error{code, err.Error()},
				Rpc:   request.Rpc,
			})
			continue
		}

		_ = connection.WriteJson(&RpcResultResponse{request.Rpc, result})
	}

	_ = socket.Close()
	connections.UnregisterConnection(connection)
	printer_objects.Unsubscribe(connection)
}
