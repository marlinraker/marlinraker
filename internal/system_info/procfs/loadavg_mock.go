//go:build !linux

package procfs

func GetLoadAvg() ([3]float32, error) {
	return [3]float32{0, 0, 0}, nil
}
