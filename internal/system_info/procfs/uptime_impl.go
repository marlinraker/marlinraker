//go:build linux

package procfs

import (
	"os"
	"strconv"
	"strings"
)

func GetUptime() (float64, error) {

	uptimeBytes, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	secondsStr := strings.Split(strings.TrimSpace(string(uptimeBytes)), " ")[0]
	seconds, err := strconv.ParseFloat(secondsStr, 64)
	if err != nil {
		return 0, err
	}
	return seconds, nil
}
