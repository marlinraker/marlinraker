package procfs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func getProcessTimeImpl(procPidStatPath string, clk int32) (float32, error) {
	statBytes, err := os.ReadFile(procPidStatPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read %q: %w", procPidStatPath, err)
	}

	parts := strings.Split(strings.TrimSpace(string(statBytes)), " ")
	if len(parts) < 15 {
		return 0, fmt.Errorf("malformed %s", procPidStatPath)
	}

	utime, err := strconv.ParseInt(parts[13], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse utime: %w", err)
	}

	stime, err := strconv.ParseInt(parts[14], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse stime: %w", err)
	}

	return float32(utime+stime) / float32(clk), nil
}
