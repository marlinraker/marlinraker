package api

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
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

func handleHttp(writer http.ResponseWriter, request *http.Request) error {

	writer.Header().Set("Content-Type", "application/json")

	method := request.Method
	url := strings.TrimRight(request.URL.Path, "/")

	params := make(executors.Params)
	values := request.URL.Query()
	for param, value := range values {
		params[param] = value[0]
	}

	if request.Body != nil && request.ContentLength > 0 {
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bodyBytes, &params); err != nil {
			return err
		}
	}

	executor := httpExecutors[method][url]
	if executor == nil {
		log.Errorln("No executor found for " + url)
		writer.WriteHeader(404)
		bytes, err := json.Marshal(&Error{Code: 404, Message: "Not Found"})
		if err != nil {
			return err
		}
		if _, err = writer.Write(bytes); err != nil {
			return err
		}
		return nil
	}

	result, err := executor(nil, request, params)
	return writeExecutorResponse(writer, result, err)
}

func writeExecutorResponse(writer http.ResponseWriter, result any, err error) error {

	if err != nil {
		log.Error(err)
		code := 500
		if executorError, isExecutorError := err.(*util.ExecutorError); isExecutorError {
			code = executorError.Code
		}
		bytes, err := json.Marshal(ErrorResponse{Error: Error{Code: code, Message: err.Error()}})
		if err != nil {
			return err
		}
		writer.WriteHeader(code)
		if _, err = writer.Write(bytes); err != nil {
			return err
		}
		return nil
	}

	var bytes []byte
	switch result.(type) {
	case files.FileUploadAction:
		bytes, err = json.Marshal(result)
	default:
		bytes, err = json.Marshal(ResultResponse{result})
	}
	if err != nil {
		return err
	}

	writer.WriteHeader(200)
	_, err = writer.Write(bytes)
	return err
}

func handleFileDownload(writer http.ResponseWriter, request *http.Request) {
	path := strings.TrimPrefix(request.URL.Path, "/server/files/")
	root := strings.Split(path, "/")[0]
	file := strings.TrimPrefix(filepath.Clean(path[len(root):]), "/")
	diskPath := filepath.Join(files.DataDir, root, file)
	http.ServeFile(writer, request, diskPath)
}

func handleFileDelete(writer http.ResponseWriter, request *http.Request) error {
	path := strings.TrimPrefix(request.URL.Path, "/server/files/")
	root := strings.Split(path, "/")[0]
	file := strings.TrimPrefix(filepath.Clean(path[len(root):]), "/")
	result, err := files.DeleteFile(root, file)
	return writeExecutorResponse(writer, result, err)
}
