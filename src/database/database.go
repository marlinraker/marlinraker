package database

import (
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/tidwall/gjson"
	"marlinraker-go/src/files"
	"os"
	"path/filepath"
)

var (
	reservedNamespaces = []string{"marlinraker", "moonraker", "gcode_metadata", "history"}
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

func GetItem(namespace string, key string) string {
	path := joinPath(namespace, key)
	result := gjson.Get(json, path)
	return result.String()
}

func ListNamespaces() []string {
	result := gjson.Get(json, "@this")
	return lo.Keys(result.Map())
}

func joinPath(namespace string, key string) string {
	if key != "" {
		return namespace + "." + key
	} else {
		return namespace
	}
}
