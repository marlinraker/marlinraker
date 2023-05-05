package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

type ServerFilesZipResult files.ZipAction

func ServerFilesZip(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	dest, _ := params.GetString("dest")

	items, exist := params.GetStringSlice("items")
	if !exist {
		return nil, util.NewError("items param is required", 400)
	}

	storeOnly, _ := params.GetBool("store_only")

	return files.CreateArchive(dest, items, !storeOnly)
}
