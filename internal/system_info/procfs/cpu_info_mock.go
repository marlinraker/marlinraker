//go:build !linux

package procfs

import (
	"runtime"
)

func GetCpuInfo() (*CpuInfo, error) {
	return &CpuInfo{
		CpuCount:     runtime.NumCPU(),
		Bits:         "",
		Processor:    "Unknown",
		CpuDesc:      "",
		SerialNumber: "Unknown",
		HardwareDesc: "",
		Model:        "Unknown",
		TotalMemory:  0,
		MemoryUnits:  "B",
	}, nil
}

func GetCpuTimes() (CpuTimes, error) {
	return CpuTimes{}, nil
}

func GetCpuTemp() (float64, error) {
	return 0, nil
}
