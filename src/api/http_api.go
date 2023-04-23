package api

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"marlinraker-go/src/api/executors"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Error Error `json:"error"`
}

type ResultResponse struct {
	Result any `json:"result"`
}

func handleHttp(writer http.ResponseWriter, request *http.Request) {

	writer.Header().Set("Content-Type", "application/json")

	method := request.Method
	url := strings.TrimRight(request.URL.Path, "/")

	params := make(executors.Params)
	values := request.URL.Query()
	for param, value := range values {
		params[param] = value[0]
	}

	executor := httpExecutors[method][url]
	if executor == nil {
		log.Errorln("No executor found for " + url)
		writer.WriteHeader(404)
		bytes, _ := json.Marshal(&Error{Code: 404, Message: "Not Found"})
		_, _ = writer.Write(bytes)
		return
	}

	result, err := executor(nil, params)
	if err != nil {
		log.Error(err)
		code := 500
		if executorError, isExecutorError := err.(*executors.ExecutorError); isExecutorError {
			code = executorError.Code
		}
		writer.WriteHeader(code)
		bytes, _ := json.Marshal(ErrorResponse{Error: Error{Code: code, Message: err.Error()}})
		_, _ = writer.Write(bytes)
		return
	}

	writer.WriteHeader(200)
	bytes, _ := json.Marshal(ResultResponse{Result: result})
	_, _ = writer.Write(bytes)
}
