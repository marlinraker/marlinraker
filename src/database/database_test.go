package database

import (
	"github.com/spf13/afero"
	"gotest.tools/assert"
	"marlinraker/src/files"
	"testing"
)

func TestInit(t *testing.T) {

	files.Fs = afero.NewMemMapFs()

	err := Init()
	assert.NilError(t, err)

	content, err := afero.ReadFile(files.Fs, "db.json")
	assert.NilError(t, err)
	assert.Equal(t, string(content), "{}")

	err = afero.WriteFile(files.Fs, "db.json", []byte("{invalid:json}"), 0755)
	assert.NilError(t, err)

	err = Init()
	assert.Error(t, err, "malformed db.json")
}

func TestPostItem(t *testing.T) {

	files.Fs = afero.NewMemMapFs()

	err := Init()
	assert.NilError(t, err)

	val, err := PostItem("test", "key", "value", false)
	assert.NilError(t, err)
	assert.Equal(t, val, "value")

	val, err = PostItem("test", "number", 123, false)
	assert.NilError(t, err)
	assert.Equal(t, val, 123)

	val, err = PostItem("test", "bool", true, false)
	assert.NilError(t, err)
	assert.Equal(t, val, true)

	val, err = PostItem("test", "slice", []string{"a", "b", "c"}, false)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, []string{"a", "b", "c"})

	val, err = PostItem("marlinraker", "foo", "bar", true)
	assert.NilError(t, err)
	assert.Equal(t, val, "bar")

	_, err = PostItem("marlinraker", "boo", "far", false)
	assert.Error(t, err, "reserved namespace access not allowed")

	content, err := afero.ReadFile(files.Fs, "db.json")
	assert.NilError(t, err)

	assert.Equal(t, string(content), `{"test":{"key":"value","number":123,"bool":true,"slice":["a","b","c"]},"marlinraker":{"foo":"bar"}}`)
}

func TestGetItem(t *testing.T) {

	files.Fs = afero.NewMemMapFs()

	content := `{"test":{"key":"value","number":123,"bool":true,"slice":["a","b","c"]},"marlinraker":{"foo":"bar"}}`
	err := afero.WriteFile(files.Fs, "db.json", []byte(content), 0755)
	assert.NilError(t, err)

	err = Init()
	assert.NilError(t, err)

	val, err := GetItem("test", "key", false)
	assert.NilError(t, err)
	assert.Equal(t, val, "value")

	val, err = GetItem("test", "number", false)
	assert.NilError(t, err)
	assert.Equal(t, val, 123.)

	val, err = GetItem("test", "bool", false)
	assert.NilError(t, err)
	assert.Equal(t, val, true)

	val, err = GetItem("test", "slice", false)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, []any{"a", "b", "c"})

	val, err = GetItem("test", "", false)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, map[string]any{
		"bool":   true,
		"key":    "value",
		"number": 123.,
		"slice":  []any{"a", "b", "c"},
	})

	_, err = GetItem("test", "foo", false)
	assert.Error(t, err, `key "foo" in namespace "test" not found`)

	_, err = GetItem("foo", "", false)
	assert.Error(t, err, `namespace "foo" not found`)

	_, err = GetItem("marlinraker", "foo", false)
	assert.Error(t, err, "reserved namespace access not allowed")

	val, err = GetItem("marlinraker", "foo", true)
	assert.NilError(t, err)
	assert.Equal(t, val, "bar")
}

func TestDeleteItem(t *testing.T) {

	files.Fs = afero.NewMemMapFs()

	{
		content := `{"test":{"key":"value","number":123,"bool":true,"slice":["a","b","c"]},"foo":{"bar":0}}`
		err := afero.WriteFile(files.Fs, "db.json", []byte(content), 0755)
		assert.NilError(t, err)
	}

	err := Init()
	assert.NilError(t, err)

	val, err := DeleteItem("test", "key", false)
	assert.NilError(t, err)
	assert.Equal(t, val, "value")

	val, err = DeleteItem("test", "number", false)
	assert.NilError(t, err)
	assert.Equal(t, val, 123.)

	val, err = DeleteItem("test", "foo", false)
	assert.Error(t, err, `key "foo" in namespace "test" not found`)

	val, err = DeleteItem("foo", "bar", false)
	assert.NilError(t, err)
	assert.Equal(t, val, 0.)

	val, err = DeleteItem("marlinraker", "foo", false)
	assert.Error(t, err, "reserved namespace access not allowed")

	{
		content, err := afero.ReadFile(files.Fs, "db.json")
		assert.NilError(t, err)
		assert.Equal(t, string(content), `{"test":{"bool":true,"slice":["a","b","c"]}}`)
	}
}

func TestListNamespaces(t *testing.T) {

	files.Fs = afero.NewMemMapFs()

	content := `{"test":{"key":"value","number":123,"bool":true,"slice":["a","b","c"]},"foo":{"bar":0}}`
	err := afero.WriteFile(files.Fs, "db.json", []byte(content), 0755)
	assert.NilError(t, err)

	err = Init()
	assert.NilError(t, err)

	namespaces := ListNamespaces()
	assert.DeepEqual(t, namespaces, append(ReservedNamespaces, "test", "foo"))
}
