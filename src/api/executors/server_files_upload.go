package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
	"path/filepath"
	"strconv"
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
		path = util.SanitizePath(values[0])
	}

	checksum := ""
	if values := form.Value["checksum"]; len(values) > 0 {
		checksum = values[0]
	}

	startPrint := false
	if values := form.Value["print"]; len(values) > 0 {
		startPrint = values[0] == "true"
	}

	headers, exist := form.File["file"]
	if !exist {
		return nil, util.NewError("cannot find file", 400)
	}

	if len(headers) > 1 {
		return nil, util.NewError("only single file upload allowed", 400)
	}

	action, err := files.Upload(root, path, checksum, headers[0])
	if err != nil {
		return nil, err
	}

	if startPrint {
		printStarted := false
		if marlinraker.Printer != nil {
			fileName := filepath.Join(path, headers[0].Filename)
			printStarted = marlinraker.Printer.PrintManager.CanPrint(fileName)
			if printStarted {
				<-marlinraker.Printer.MainExecutorContext().QueueGcode("SDCARD_PRINT_FILE FILENAME="+strconv.Quote(fileName), true)
			}
		}
		action.PrintStarted = &printStarted
	}
	return action, nil
}
