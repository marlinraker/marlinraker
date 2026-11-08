package procfs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func getLoadAvgImpl(loadAvgPath string) ([3]float32, error) {
	var loadAvg [3]float32
	loadAvgBytes, err := os.ReadFile(loadAvgPath)
	if err != nil {
		return loadAvg, fmt.Errorf("failed to read %q: %w", loadAvgPath, err)
	}

	parts := strings.Split(strings.TrimSpace(string(loadAvgBytes)), " ")
	for i := 0; i < 3; i++ {
		avg, err := strconv.ParseFloat(parts[i], 32)
		if err != nil {
			return loadAvg, fmt.Errorf("failed to parse load average: %w", err)
		}
		loadAvg[i] = float32(avg)
	}
	return loadAvg, nil
}
