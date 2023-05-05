package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

type ServerFilesUploadResult files.FileUploadAction

func ServerFilesUpload(_ *connections.Connection, request *http.Request, _ Params) (any, error) {

	reader, err := request.MultipartReader()
	if err != nil {
		return nil, err
	}

	form, err := reader.ReadForm(8 << 20) // 8 mb
	if err != nil {
		return nil, err
	}

	root := "gcodes"
	if values := form.Value["root"]; len(values) > 0 {
		root = values[0]
	}
	if root != "gcodes" && root != "config" {
		return nil, util.NewError("unallowed root \""+root+"\"", 400)
	}

	path := ""
	if values := form.Value["path"]; len(values) > 0 {
		path = values[0]
	}

	checksum := ""
	if values := form.Value["checksum"]; len(values) > 0 {
		checksum = values[0]
	}

	headers, exist := form.File["file"]
	if !exist {
		return nil, util.NewError("cannot find file", 400)
	}

	if len(headers) > 1 {
		return nil, util.NewError("only single file upload allowed", 400)
	}

	return files.Upload(root, path, checksum, headers[0])
}
