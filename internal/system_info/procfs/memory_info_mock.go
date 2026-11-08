//go:build !linux

package procfs

func GetUsedMem() (int64, string, error) {
	return 0, "B", nil
}

func GetMemAvail() (int64, error) {
	return 0, nil
}
