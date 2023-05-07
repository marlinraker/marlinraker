package database

import (
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/tidwall/gjson"
	"marlinraker/src/files"
	"marlinraker/src/util"
	"os"
	"path/filepath"
)

var (
	ReservedNamespaces = []string{"marlinraker", "moonraker", "gcode_metadata", "history"}
	dbFile             string
	json               string
)

func Init() error {
	dbFile = filepath.Join(files.DataDir, "db.json")
	_, err := files.Fs.OpenFile(dbFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	jsonBytes, err := afero.ReadFile(files.Fs, dbFile)
	if err != nil {
		return err
	}
	json = string(jsonBytes)
	return nil
}

func GetItem(namespace string, key string) (any, error) {
	path := joinPath(namespace, key)
	result := gjson.Get(json, path).Value()
	if result == nil {
		if key != "" {
			return nil, util.NewError("Key \""+key+"\" in namespace \""+namespace+"\" not found", 404)
		} else {
			return nil, util.NewError("Namespace \""+namespace+"\" not found", 404)
		}
	}
	return result, nil
}

func ListNamespaces() []string {
	result := gjson.Get(json, "@this")
	return append(ReservedNamespaces, lo.Keys(result.Map())...)
}

func joinPath(namespace string, key string) string {
	if key != "" {
		return namespace + "." + key
	} else {
		return namespace
	}
}
