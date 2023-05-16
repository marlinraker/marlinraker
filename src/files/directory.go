package files

import (
	"errors"
	"github.com/spf13/afero"
	"marlinraker/src/api/notification"
	"path/filepath"
	"strings"
	"syscall"
)

type RootInfo struct {
	Name        string `json:"name"`
	Permissions string `json:"permissions"`
}

type DiskUsage struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

type DirectoryMeta struct {
	Modified    float64 `json:"modified"`
	Size        int64   `json:"size"`
	Permissions string  `json:"permissions"`
	DirName     string  `json:"dirname"`
}

type DirectoryInfo struct {
	Dirs      []DirectoryMeta `json:"dirs"`
	Files     []any           `json:"files"`
	DiskUsage DiskUsage       `json:"disk_usage"`
	RootInfo  RootInfo        `json:"root_info"`
}

type DirectoryAction struct {
	Item   ActionItem `json:"item"`
	Action string     `json:"action"`
}

func GetDirInfo(path string, extended bool) (DirectoryInfo, error) {

	parts := strings.Split(path, "/")
	rootName := parts[0]

	root, err := getRootByName(rootName)
	if err != nil {
		return DirectoryInfo{}, err
	}
	rootInfo := RootInfo{root.Name, root.Permissions}

	diskPath := filepath.Join(DataDir, path)
	dirContent, err := afero.ReadDir(Fs, diskPath)
	if err != nil {
		return DirectoryInfo{}, err
	}

	dirs, files := make([]DirectoryMeta, 0), make([]any, 0)
	for _, file := range dirContent {
		if file.IsDir() {
			dirs = append(dirs, DirectoryMeta{
				Modified:    float64(file.ModTime().UnixMilli()) / 1000.0,
				Size:        file.Size(),
				Permissions: root.Permissions,
				DirName:     file.Name(),
			})

		} else {
			ext := filepath.Ext(file.Name())
			if root.Name == "gcodes" && (ext == ".meta" || ext == ".png") {
				continue
			}

			var metadata *Metadata
			if extended && root.Name == "gcodes" {
				relPath, err := filepath.Rel("gcodes/", filepath.Join(path, file.Name()))
				if err != nil {
					return DirectoryInfo{}, err
				}
				if metadata, _ = LoadOrScanMetadata(relPath); metadata != nil {
					files = append(files, ExtendedFileMeta{metadata, root.Permissions})
					continue
				}
			}

			fileMeta := FileMeta{
				Modified:    float64(file.ModTime().UnixMilli()) / 1000.0,
				Size:        file.Size(),
				Permissions: root.Permissions,
				FileName:    file.Name(),
			}
			files = append(files, fileMeta)
		}
	}

	var stat syscall.Statfs_t
	err = syscall.Statfs(diskPath, &stat)
	if err != nil {
		return DirectoryInfo{}, err
	}
	total := uint64(stat.Bsize) * stat.Blocks
	free := uint64(stat.Bsize) * stat.Bfree
	used := total - free
	diskUsage := DiskUsage{total, used, free}

	return DirectoryInfo{
		Dirs:      dirs,
		Files:     files,
		DiskUsage: diskUsage,
		RootInfo:  rootInfo,
	}, nil
}

func CreateDir(path string) (DirectoryAction, error) {

	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	rootName, relPath := parts[0], strings.Join(parts[1:], "/")

	root, err := getRootByName(rootName)
	action := DirectoryAction{Action: "create_dir"}
	if err != nil {
		return action, err
	}

	if !strings.Contains(root.Permissions, "w") {
		return action, errors.New("no write permissions")
	}

	diskPath := filepath.Join(DataDir, path)
	err = Fs.Mkdir(diskPath, 0755)
	if err != nil {
		return action, err
	}

	stat, err := Fs.Stat(diskPath)
	if err != nil {
		return action, err
	}

	action.Item = ActionItem{
		Path:        relPath,
		Root:        root.Name,
		Modified:    float64(stat.ModTime().UnixMilli()) / 1000.0,
		Size:        stat.Size(),
		Permissions: root.Permissions,
	}

	err = notification.Publish(notification.New("notify_filelist_changed", []any{action}))
	return action, err
}

func DeleteDir(path string, force bool) (DirectoryAction, error) {

	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	rootName, relPath := parts[0], strings.Join(parts[1:], "/")

	action := DirectoryAction{Action: "delete_dir"}
	root, err := getRootByName(rootName)
	if err != nil {
		return action, err
	}

	diskPath := filepath.Join(DataDir, path)
	stat, err := Fs.Stat(diskPath)
	if err != nil {
		return action, err
	}

	if !stat.IsDir() {
		return action, errors.New("\"" + path + "\" is not a directory")
	}
	if !strings.Contains(root.Permissions, "w") {
		return action, errors.New("no write permissions")
	}

	files, err := afero.ReadDir(Fs, diskPath)
	if err != nil {
		return action, err
	}
	if !force && len(files) > 0 {
		return action, errors.New("directory is not empty")
	}

	err = Fs.RemoveAll(diskPath)
	if err != nil {
		return action, err
	}

	action.Item = ActionItem{
		Path:        relPath,
		Root:        root.Name,
		Modified:    float64(stat.ModTime().UnixMilli()) / 1000.0,
		Size:        stat.Size(),
		Permissions: root.Permissions,
	}

	err = notification.Publish(notification.New("notify_filelist_changed", []any{action}))
	return action, err
}
