package webcams

import (
	"testing"

	"github.com/spf13/afero"
	"gotest.tools/v3/assert"

	"github.com/marlinraker/marlinraker/internal/api/notification"
	"github.com/marlinraker/marlinraker/internal/database"
	"github.com/marlinraker/marlinraker/internal/files"
)

func TestWebcams(t *testing.T) {
	notification.Testing = true
	files.Fs = afero.NewMemMapFs()
	files.DataDir = "/data"
	assert.NilError(t, database.Init())

	list, err := List()
	assert.NilError(t, err)
	assert.Equal(t, len(list), 0)

	_, err = PostItem(map[string]any{})
	assert.ErrorContains(t, err, "required")

	created, err := PostItem(map[string]any{"name": "cam", "stream_url": "/webcam/?action=stream", "rotation": 90., "source": "config"})
	assert.NilError(t, err)
	uid, _ := created["uid"].(string)
	assert.Assert(t, uid != "")
	assert.Equal(t, created["source"], "database")
	assert.Equal(t, created["rotation"], 90.)
	assert.Equal(t, created["service"], "mjpegstreamer-adaptive")

	// update by name keeps the uid
	updated, err := PostItem(map[string]any{"name": "cam", "enabled": false})
	assert.NilError(t, err)
	assert.Equal(t, updated["uid"], uid)
	assert.Equal(t, updated["enabled"], false)
	assert.Equal(t, updated["stream_url"], "/webcam/?action=stream")

	// rename by uid
	renamed, err := PostItem(map[string]any{"uid": uid, "name": "front"})
	assert.NilError(t, err)
	assert.Equal(t, renamed["name"], "front")

	_, err = PostItem(map[string]any{"uid": "unknown"})
	assert.ErrorContains(t, err, "not found")

	_, err = PostItem(map[string]any{"name": "back"})
	assert.NilError(t, err)
	_, err = PostItem(map[string]any{"uid": uid, "name": "back"})
	assert.ErrorContains(t, err, "already exists")

	got, err := GetItem("", "front")
	assert.NilError(t, err)
	assert.Equal(t, got["uid"], uid)

	deleted, err := DeleteItem(uid, "")
	assert.NilError(t, err)
	assert.Equal(t, deleted["name"], "front")

	list, err = List()
	assert.NilError(t, err)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0]["name"], "back")

	_, err = DeleteItem("", "front")
	assert.ErrorContains(t, err, "not found")
}
