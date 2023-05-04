package executors

import (
	"fmt"
	"github.com/samber/lo"
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/util"
	"net/http"
)

type ServerFilesZipResult files.ZipAction

func ServerFilesZip(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	dest, _ := params["dest"].(string)

	itemsParam, exist := params["items"].([]any)
	if !exist {
		return nil, util.NewError("items param is required", 400)
	}
	items := lo.Map(itemsParam, func(item any, _ int) string { return fmt.Sprintf("%v", item) })

	storeOnlyParam := params["store_only"]
	storeOnly := storeOnlyParam == true || storeOnlyParam == "true"

	return files.CreateArchive(dest, items, !storeOnly)
}
