package procfs

import (
	"os"
	"strconv"
	"strings"
)

func GetLoadAvg() ([3]float32, error) {
	return getLoadAvgImpl("/proc/loadavg")
}

func getLoadAvgImpl(loadAvgPath string) ([3]float32, error) {
	var loadAvg [3]float32
	loadAvgBytes, err := os.ReadFile(loadAvgPath)
	if err != nil {
		return loadAvg, err
	}

	parts := strings.Split(strings.TrimSpace(string(loadAvgBytes)), " ")
	for i := 0; i < 3; i++ {
		avg, err := strconv.ParseFloat(parts[i], 32)
		if err != nil {
			return loadAvg, err
		}
		loadAvg[i] = float32(avg)
	}
	return loadAvg, nil
}
