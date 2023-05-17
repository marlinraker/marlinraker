package util

import "path/filepath"

func StringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func SanitizePath(path string) string {
	// prevent access outside of current folder
	// dir/../../../dir/file -> dir/file
	return filepath.Join("/", path)[1:]
}
