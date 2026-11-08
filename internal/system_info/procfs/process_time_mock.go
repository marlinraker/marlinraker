//go:build !linux

package procfs

func GetProcessTime() (float32, error) {
	return 0, nil
}
