package api

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"marlinraker/src/api/executors"
	"marlinraker/src/files"
	"marlinraker/src/util"
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

	if request.Body != nil && request.ContentLength > 0 && request.Header.Get("Content-Type") == "application/json" {
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
	return writeExecutorResponse(writer, method, url, result, err)
}

func writeExecutorResponse(writer http.ResponseWriter, method string, url string, result any, err error) error {

	if err != nil {
		log.Errorln("Error while executing "+method+" "+url+":", err)
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
	path := util.SanitizePath(strings.TrimPrefix(request.URL.Path, "/server/files/"))
	diskPath := filepath.Join(files.DataDir, util.SanitizePath(path))
	http.ServeFile(writer, request, diskPath)
}

func handleFileDelete(writer http.ResponseWriter, request *http.Request) error {
	path := strings.TrimPrefix(request.URL.Path, "/server/files/")
	result, err := files.DeleteFile(util.SanitizePath(path))
	return writeExecutorResponse(writer, request.Method, request.URL.Path, result, err)
}
