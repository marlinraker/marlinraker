package api

import (
	"marlinraker/src/files"
	"testing"
)

func TestFileRootOf(t *testing.T) {
	oldRoots := files.FileRoots
	files.FileRoots = []files.FileRoot{{Name: "gcodes"}, {Name: "config"}, {Name: "logs"}}
	t.Cleanup(func() { files.FileRoots = oldRoots })

	tests := map[string]string{
		"/server/files/gcodes/benchy.gcode":       "gcodes",
		"/server/files/config/sub/printer.toml":   "config",
		"/server/files/gcodes/../config/x.toml":   "config",
		"/server/files/gcodes/../db.json":         "",
		"/server/files/gcodes/../../etc/passwd":   "",
		"/server/files/gcodesfoo/x":               "",
		"/server/files/gcodes/../../../gcodes/xy": "gcodes",
		"/server/info": "",
	}
	for path, want := range tests {
		root, ok := fileRootOf(path)
		if ok != (want != "") || root.Name != want {
			t.Errorf("fileRootOf(%q) = %q, %v; want %q", path, root.Name, ok, want)
		}
	}
}
