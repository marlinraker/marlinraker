package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerFilesZipResult files.ZipAction

func ServerFilesZip(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	dest, _ := params.RequirePath("dest")

	items, err := params.RequireStringSlice("items")
	if err != nil {
		return nil, err
	}

	storeOnly, _ := params.GetBool("store_only")

	return files.CreateArchive(dest, items, !storeOnly)
}
