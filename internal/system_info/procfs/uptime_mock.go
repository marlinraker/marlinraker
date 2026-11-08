//go:build !linux

package procfs

func GetUptime() (float64, error) {
	return 0, nil
}
