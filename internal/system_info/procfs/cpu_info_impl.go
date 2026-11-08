//go:build linux

package procfs

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
)

func GetCpuInfo() (*CpuInfo, error) {
	return getCpuInfoImpl("/proc/cpuinfo", "/proc/meminfo")
}

func GetCpuTimes() (CpuTimes, error) {
	return getCpuTimesImpl("/proc/stat")
}

var cpuTempAvail = true

func GetCpuTemp() (float64, error) {
	if !cpuTempAvail {
		return 0, nil
	}
	temp, err := getCpuTempImpl("/sys/class/thermal/thermal_zone0/temp")
	if err != nil && errors.Is(err, os.ErrNotExist) {
		log.Warn("Temperature readings not available at /sys/class/thermal/thermal_zone0/temp. Host temperature will not be displayed.")
		cpuTempAvail = false
		return 0, nil
	}
	return temp, err
}
