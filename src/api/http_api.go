package api

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"marlinraker-go/src/api/executors"
	"marlinraker-go/src/files"
	"marlinraker-go/src/util"
	"net/http"
	"path/filepath"
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
		bytes, err := json.Marshal(&Error{Code: 404, Message: "Not Found"})
		if err != nil {
			log.Error(err)
			return
		}
		if _, err = writer.Write(bytes); err != nil {
			log.Error(err)
		}
		return
	}

	result, err := executor(nil, request, params)
	writeExecutorResponse(writer, result, err)
}

func writeExecutorResponse(writer http.ResponseWriter, result any, err error) {

	if err != nil {
		log.Error(err)
		code := 500
		if executorError, isExecutorError := err.(*util.ExecutorError); isExecutorError {
			code = executorError.Code
		}
		writer.WriteHeader(code)
		bytes, err := json.Marshal(ErrorResponse{Error: Error{Code: code, Message: err.Error()}})
		if err != nil {
			log.Error(err)
			return
		}
		if _, err = writer.Write(bytes); err != nil {
			log.Error(err)
		}
		return
	}

	var bytes []byte
	switch result.(type) {
	case files.FileUploadAction:
		bytes, err = json.Marshal(result)
	default:
		bytes, err = json.Marshal(ResultResponse{result})
	}
	if err != nil {
		log.Error(err)
		return
	}

	writer.WriteHeader(200)
	_, _ = writer.Write(bytes)
}

func handleFileDownload(writer http.ResponseWriter, request *http.Request) {
	path := strings.TrimPrefix(request.URL.Path, "/server/files/")
	root := strings.Split(path, "/")[0]
	file := strings.TrimPrefix(filepath.Clean(path[len(root):]), "/")
	diskPath := filepath.Join(files.DataDir, root, file)
	http.ServeFile(writer, request, diskPath)
}

func handleFileDelete(writer http.ResponseWriter, request *http.Request) {
	path := strings.TrimPrefix(request.URL.Path, "/server/files/")
	root := strings.Split(path, "/")[0]
	file := strings.TrimPrefix(filepath.Clean(path[len(root):]), "/")
	result, err := files.DeleteFile(root, file)
	writeExecutorResponse(writer, result, err)
}
