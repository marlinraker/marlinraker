//go:build linux

package procfs

func GetUsedMem() (int64, string, error) {
	return getUsedMemImpl("/proc/meminfo")
}

func GetMemAvail() (int64, error) {
	return getMemAvailImpl("/proc/meminfo")
}
