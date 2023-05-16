package files

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"marlinraker/src/api/notification"
	"marlinraker/src/util"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type FileMeta struct {
	Modified    float64 `json:"modified"`
	Size        int64   `json:"size"`
	Permissions string  `json:"permissions"`
	FileName    string  `json:"filename"`
}

type ExtendedFileMeta struct {
	*Metadata
	Permissions string `json:"permissions"`
}

type FileUploadAction struct {
	Item   ActionItem `json:"item"`
	Action string     `json:"action"`
}

type FileDeleteAction struct {
	Item   ActionItem `json:"item"`
	Action string     `json:"action"`
}

func Upload(rootName string, path string, checksum string, header *multipart.FileHeader) (FileUploadAction, error) {

	fileName := header.Filename
	sourceFile, err := header.Open()
	if err != nil {
		return FileUploadAction{}, err
	}

	defer func() {
		if err := sourceFile.Close(); err != nil {
			util.LogError(err)
		}
	}()

	root, err := getRootByName(rootName)
	if err != nil {
		return FileUploadAction{}, err
	}

	if !strings.Contains(root.Permissions, "w") {
		return FileUploadAction{}, errors.New("no write permissions")
	}

	destDirPath := filepath.Join(root.Path, path)
	destPath := filepath.Join(destDirPath, fileName)

	if err := Fs.MkdirAll(destDirPath, 0755); err != nil {
		return FileUploadAction{}, err
	}

	_, err = Fs.Stat(destPath)
	existedBefore := err == nil

	destFile, err := Fs.OpenFile(destPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return FileUploadAction{}, err
	}

	defer func() {
		if err := destFile.Close(); err != nil {
			util.LogError(err)
		}
	}()

	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return FileUploadAction{}, err
	}

	if checksum != "" {
		hash, err := calculateChecksum(destPath)
		if err != nil {
			return FileUploadAction{}, err
		}
		if checksum != hash {
			return FileUploadAction{}, util.NewError("checksums do not match (got "+checksum+", expected: "+hash+")", 422)
		}
	}

	stat, err := Fs.Stat(destPath)
	if err != nil {
		return FileUploadAction{}, err
	}

	actionName := "create_file"
	if existedBefore {
		actionName = "modify_file"
	}

	action := FileUploadAction{
		Item: ActionItem{
			Path:        filepath.Join(path, fileName),
			Root:        root.Name,
			Modified:    float64(stat.ModTime().UnixMilli()) / 1000.0,
			Size:        stat.Size(),
			Permissions: root.Permissions,
		},
		Action: actionName,
	}

	err = notification.Publish(notification.New("notify_filelist_changed", []any{action}))
	return action, err
}

func DeleteFile(rootName string, fileName string) (FileDeleteAction, error) {

	root, err := getRootByName(rootName)
	if err != nil {
		return FileDeleteAction{}, err
	}

	if !strings.Contains(root.Permissions, "w") {
		return FileDeleteAction{}, errors.New("no write permissions")
	}

	diskPath := filepath.Join(DataDir, rootName, fileName)
	stat, err := Fs.Stat(diskPath)
	if err != nil {
		return FileDeleteAction{}, err
	}
	if stat.IsDir() {
		return FileDeleteAction{}, util.NewError(diskPath+" is a directory", 400)
	}
	if err := Fs.Remove(diskPath); err != nil {
		return FileDeleteAction{}, err
	}

	if HasMetadata(fileName) {
		if err := RemoveMetadata(fileName); err != nil {
			return FileDeleteAction{}, err
		}
	}

	action := FileDeleteAction{
		Item: ActionItem{
			Path:        fileName,
			Root:        rootName,
			Size:        0,
			Modified:    0,
			Permissions: "",
		},
		Action: "delete_file",
	}

	err = notification.Publish(notification.New("notify_filelist_changed", []any{action}))
	return action, err
}

func calculateChecksum(filePath string) (string, error) {

	file, err := Fs.Open(filePath)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := file.Close(); err != nil {
			util.LogError(err)
		}
	}()

	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}

	checksum := fmt.Sprintf("%x", digest.Sum(nil))
	return checksum, nil
}
