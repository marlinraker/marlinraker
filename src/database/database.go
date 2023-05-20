package database

import (
	"errors"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"marlinraker/src/files"
	"marlinraker/src/util"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ReservedNamespaces = []string{"marlinraker", "moonraker", "gcode_metadata", "history"}
	dbFile             string
	json               string
	mu                 = &sync.RWMutex{}
)

func Init() error {

	dbFile = filepath.Join(files.DataDir, "db.json")
	file, err := files.Fs.OpenFile(dbFile, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}

	jsonBytes, err := afero.ReadAll(file)
	if err != nil {
		return err
	}

	if len(jsonBytes) == 0 {
		jsonBytes = []byte("{}")
		if _, err := file.Write(jsonBytes); err != nil {
			return err
		}
	}

	if !gjson.ValidBytes(jsonBytes) {
		return errors.New("malformed db.json")
	}

	json = string(jsonBytes)
	return nil
}

func GetItem(namespace string, key string, internal bool) (any, error) {
	mu.RLock()
	defer mu.RUnlock()
	return getItem(namespace, key, internal)
}

func getItem(namespace string, key string, internal bool) (any, error) {

	if !internal && lo.Contains(ReservedNamespaces, strings.ToLower(namespace)) {
		return nil, errors.New("reserved namespace access not allowed")
	}

	path := joinPath(namespace, key)
	result := gjson.Get(json, path).Value()

	if result == nil {
		if key != "" {
			return nil, util.NewError("key \""+key+"\" in namespace \""+namespace+"\" not found", 404)
		} else {
			return nil, util.NewError("namespace \""+namespace+"\" not found", 404)
		}
	}
	return result, nil
}

func PostItem(namespace string, key string, value any, internal bool) (any, error) {

	if !internal && lo.Contains(ReservedNamespaces, strings.ToLower(namespace)) {
		return nil, errors.New("reserved namespace access not allowed")
	}

	mu.Lock()
	defer mu.Unlock()

	path := joinPath(namespace, key)

	var err error
	json, err = sjson.Set(json, path, value)
	if err != nil {
		return nil, err
	}

	if err := afero.WriteFile(files.Fs, dbFile, []byte(json), 0755); err != nil {
		return nil, err
	}
	return value, nil
}

func DeleteItem(namespace string, key string, internal bool) (any, error) {

	mu.Lock()
	defer mu.Unlock()

	value, err := getItem(namespace, key, internal)
	if err != nil {
		return nil, err
	}

	path := joinPath(namespace, key)
	if json, err = sjson.Delete(json, path); err != nil {
		return nil, err
	}

	if len(gjson.Get(json, namespace).Map()) == 0 {
		if json, err = sjson.Delete(json, namespace); err != nil {
			return nil, err
		}
	}

	if err := afero.WriteFile(files.Fs, dbFile, []byte(json), 0755); err != nil {
		return nil, err
	}
	return value, nil
}

func ListNamespaces() []string {
	mu.RLock()
	defer mu.RUnlock()
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
