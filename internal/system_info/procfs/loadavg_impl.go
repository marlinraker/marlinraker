//go:build linux

package procfs

func GetLoadAvg() ([3]float32, error) {
	return getLoadAvgImpl("/proc/loadavg")
}
