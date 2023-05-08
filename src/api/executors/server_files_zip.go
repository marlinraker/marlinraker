package executors

import (
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type ServerFilesZipResult files.ZipAction

func ServerFilesZip(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	dest, _ := params.GetString("dest")

	items, err := params.RequireStringSlice("items")
	if err != nil {
		return nil, err
	}

	storeOnly, _ := params.GetBool("store_only")

	return files.CreateArchive(dest, items, !storeOnly)
}
