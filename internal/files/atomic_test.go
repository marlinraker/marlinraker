package files

import (
	"testing"

	"github.com/spf13/afero"
)

func TestWriteFileAtomic(t *testing.T) {
	oldFs := Fs
	Fs = afero.NewMemMapFs()
	t.Cleanup(func() { Fs = oldFs })

	if err := afero.WriteFile(Fs, "/data/db.json", []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic("/data/db.json", []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	content, err := afero.ReadFile(Fs, "/data/db.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf("content = %q, want %q", content, "new")
	}

	entries, err := afero.ReadDir(Fs, "/data")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected temporary file to be gone, found %d entries", len(entries))
	}
}
