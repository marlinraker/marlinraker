package procfs

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

var (
	totalMemRegex = regexp.MustCompile(`(?m)^MemTotal:\s*([0-9]+) ([A-Za-z]+)$`)
	freeMemRegex  = regexp.MustCompile(`(?m)^MemFree:\s*([0-9]+) ([A-Za-z]+)$`)
	memAvailRegex = regexp.MustCompile(`(?m)^MemAvailable:\s*([0-9]+) kB$`)
)

func getTotalMemImpl(memInfoPath string) (int64, string, error) {
	memInfoBytes, err := os.ReadFile(memInfoPath)
	if err != nil {
		return 0, "B", fmt.Errorf("failed to read %q: %w", memInfoPath, err)
	}
	memInfo := string(memInfoBytes)

	totalMem, unit, err := getMemory(memInfo, totalMemRegex)
	if err != nil {
		return 0, "B", fmt.Errorf("failed to get total memory: %w", err)
	}
	return totalMem, unit, nil
}

func getUsedMemImpl(memInfoPath string) (int64, string, error) {
	memInfoBytes, err := os.ReadFile(memInfoPath)
	if err != nil {
		return 0, "B", fmt.Errorf("failed to read %q: %w", memInfoPath, err)
	}
	memInfo := string(memInfoBytes)

	totalMem, unit, err := getMemory(memInfo, totalMemRegex)
	if err != nil {
		return 0, "B", fmt.Errorf("failed to get total memory: %w", err)
	}

	freeMem, _, err := getMemory(memInfo, freeMemRegex)
	if err != nil {
		return 0, "B", fmt.Errorf("failed to get free memory: %w", err)
	}

	return totalMem - freeMem, unit, nil
}

func getMemory(memInfo string, regex *regexp.Regexp) (int64, string, error) {
	if match := regex.FindStringSubmatch(memInfo); match != nil {
		totalMemory, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return 0, "B", fmt.Errorf("failed to parse memory: %w", err)
		}
		units := match[2]
		return totalMemory, units, nil
	} else {
		return 0, "B", nil
	}
}

func getMemAvailImpl(memInfoPath string) (int64, error) {
	memInfoBytes, err := os.ReadFile(memInfoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read %q: %w", memInfoPath, err)
	}

	if match := memAvailRegex.FindStringSubmatch(string(memInfoBytes)); match != nil {
		var memAvail int64
		memAvail, err = strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse meminfo: %w", err)
		}
		return memAvail, nil
	}

	return 0, errors.New("malformed meminfo")
}
