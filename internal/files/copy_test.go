package files

import (
	"testing"

	"github.com/spf13/afero"
	"gotest.tools/v3/assert"

	"github.com/marlinraker/marlinraker/internal/api/notification"
)

func setupFs(t *testing.T) {
	t.Helper()
	oldFs, oldDataDir := Fs, DataDir
	t.Cleanup(func() { Fs, DataDir = oldFs, oldDataDir })
	notification.Testing = true
	Fs, DataDir = afero.NewMemMapFs(), "/data"
	assert.NilError(t, Init())
}

func TestCopy(t *testing.T) {
	setupFs(t)
	assert.NilError(t, afero.WriteFile(Fs, "/data/gcodes/a.gcode", []byte("G28\n"), 0644))
	assert.NilError(t, afero.WriteFile(Fs, "/data/gcodes/dir/b.gcode", []byte("G1 X1\n"), 0644))
	assert.NilError(t, afero.WriteFile(Fs, "/data/logs/marlinraker.log", []byte("log"), 0644))

	action, err := Copy("gcodes/a.gcode", "gcodes/sub/copy.gcode")
	assert.NilError(t, err)
	assert.Equal(t, action.Action, "create_file")
	assert.Equal(t, action.Item.Path, "sub/copy.gcode")
	assert.Equal(t, action.Item.Size, int64(4))
	content, err := afero.ReadFile(Fs, "/data/gcodes/sub/copy.gcode")
	assert.NilError(t, err)
	assert.Equal(t, string(content), "G28\n")

	action, err = Copy("gcodes/dir", "config/dir2")
	assert.NilError(t, err)
	assert.Equal(t, action.Action, "create_dir")
	content, err = afero.ReadFile(Fs, "/data/config/dir2/b.gcode")
	assert.NilError(t, err)
	assert.Equal(t, string(content), "G1 X1\n")

	_, err = Copy("gcodes/a.gcode", "gcodes/sub/copy.gcode")
	assert.ErrorContains(t, err, "already exists")
	_, err = Copy("gcodes/missing.gcode", "gcodes/x.gcode")
	assert.ErrorContains(t, err, "does not exist")
	_, err = Copy("gcodes/a.gcode", "logs/a.gcode")
	assert.ErrorContains(t, err, "no write permissions")
	_, err = Copy("gcodes/dir", "gcodes/dir/inner")
	assert.ErrorContains(t, err, "into itself")
	_, err = Copy("gcodes/a.gcode", "gcodes")
	assert.ErrorContains(t, err, "file root")

	// reading from a read-only root is fine
	_, err = Copy("logs/marlinraker.log", "config/marlinraker.log")
	assert.NilError(t, err)
}

func TestMoveChecksPermissionsFirst(t *testing.T) {
	setupFs(t)
	assert.NilError(t, afero.WriteFile(Fs, "/data/logs/marlinraker.log", []byte("log"), 0644))

	_, err := Move("logs/marlinraker.log", "gcodes/marlinraker.log")
	assert.ErrorContains(t, err, "no write permissions")

	_, err = Fs.Stat("/data/logs/marlinraker.log")
	assert.NilError(t, err, "file must not be moved out of a read-only root")
}
