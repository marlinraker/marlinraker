package executors

import (
	"net/http"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/util"
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
	// ReadForm spills large files to temporary files on disk
	defer func() {
		if err := form.RemoveAll(); err != nil {
			log.Errorf("Failed to remove temporary upload files: %v", err)
		}
	}()

	root := "gcodes"
	if values := form.Value["root"]; len(values) > 0 {
		root = values[0]
	}
	if root != "gcodes" && root != "config" {
		return nil, util.NewErrorf(400, "unallowed root %q", root)
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
		return nil, util.NewError(400, "cannot find file")
	}

	if len(headers) > 1 {
		return nil, util.NewError(400, "only single file upload allowed")
	}

	action, err := files.Upload(root, path, checksum, headers[0])
	if err != nil {
		return nil, err
	}

	if startPrint {
		printStarted := false
		if printer := marlinraker.GetPrinter(); printer != nil {
			fileName := filepath.Join(path, headers[0].Filename)
			printStarted = printer.PrintManager.CanPrint(fileName)
			if printStarted {
				<-printer.MainExecutorContext().QueueGcode("SDCARD_PRINT_FILE FILENAME="+strconv.Quote(fileName), true)
			}
		}
		action.PrintStarted = &printStarted
	}
	return action, nil
}
