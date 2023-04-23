package files

import (
	"errors"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"io/fs"
	"marlinraker-go/src/api/notification"
	"path/filepath"
	"strings"
)

type File struct {
	Path        string  `json:"path"`
	Modified    float32 `json:"modified"`
	Size        int64   `json:"size"`
	Permissions string  `json:"permissions"`
}

type FileRoot struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Permissions string `json:"permissions"`
}

type ActionItem struct {
	Path        string  `json:"path"`
	Root        string  `json:"root"`
	Modified    float64 `json:"modified,omitempty"`
	Size        int64   `json:"size,omitempty"`
	Permissions string  `json:"permissions,omitempty"`
}

type MoveAction struct {
	Item       ActionItem `json:"item"`
	SourceItem ActionItem `json:"source_item"`
	Action     string     `json:"action"`
}

var (
	Fs        afero.Fs = &afero.OsFs{}
	DataDir   string
	FileRoots []FileRoot
)

func CreateFileRoots() error {

	FileRoots = []FileRoot{
		{
			Name:        "config",
			Path:        filepath.Join(DataDir, "config"),
			Permissions: "rw",
		},
		{
			Name:        "gcodes",
			Path:        filepath.Join(DataDir, "gcodes"),
			Permissions: "rw",
		},
		{
			Name:        "logs",
			Path:        filepath.Join(DataDir, "logs"),
			Permissions: "r",
		},
	}

	for _, fileRoot := range FileRoots {
		err := Fs.MkdirAll(fileRoot.Path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetRegisteredDirectories() []string {
	return lo.Map(FileRoots, func(root FileRoot, _ int) string { return root.Name })
}

func ListFiles(rootName string) ([]File, error) {

	root, err := getRootByName(rootName)
	if err != nil {
		return nil, err
	}

	files := make([]File, 0)

	err = afero.Walk(Fs, root.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(root.Path, path)
			if err != nil {
				return err
			}

			files = append(files, File{
				Path:        relPath,
				Modified:    float32(info.ModTime().UnixMilli()) / 1000.0,
				Size:        info.Size(),
				Permissions: root.Permissions,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return files, nil
}

func Move(source string, dest string) (MoveAction, error) {

	action := MoveAction{}
	source, dest = strings.TrimPrefix(source, "/"), strings.TrimPrefix(dest, "/")
	sourceParts, destParts := strings.Split(source, "/"), strings.Split(dest, "/")

	sourceRootName, sourceRelPath := sourceParts[0], strings.Join(sourceParts[1:], "/")
	sourceRoot, err := getRootByName(sourceRootName)
	if err != nil {
		return action, err
	}

	destRootName, destRelPath := destParts[0], strings.Join(destParts[1:], "/")
	destRoot, err := getRootByName(destRootName)
	if err != nil {
		return action, err
	}

	sourceDiskPath, destDiskPath := filepath.Join(DataDir, source), filepath.Join(DataDir, dest)
	err = Fs.Rename(sourceDiskPath, destDiskPath)
	if err != nil {
		return action, err
	}

	if !strings.Contains(sourceRoot.Permissions, "w") || !strings.Contains(destRoot.Permissions, "w") {
		return action, errors.New("no write permissions")
	}

	stat, err := Fs.Stat(destDiskPath)
	if err != nil {
		return action, err
	}

	if stat.IsDir() {
		action.Action = "move_dir"
	} else {
		action.Action = "move_file"
	}

	action.Item = ActionItem{
		Root:        destRoot.Name,
		Path:        destRelPath,
		Modified:    float64(stat.ModTime().UnixMilli()) / 1000.0,
		Size:        stat.Size(),
		Permissions: destRoot.Permissions,
	}

	action.SourceItem = ActionItem{
		Root: sourceRoot.Name,
		Path: sourceRelPath,
	}

	_ = notification.Publish(notification.New("notify_filelist_changed", []any{action}))
	return action, nil
}

func getRootByName(rootName string) (FileRoot, error) {
	root, exists := lo.Find(FileRoots, func(item FileRoot) bool {
		return item.Name == rootName
	})
	if !exists {
		return FileRoot{}, errors.New("cannot find file root \"" + rootName + "\"")
	}
	return root, nil
}
