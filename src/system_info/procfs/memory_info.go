package procfs

import (
	"errors"
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
		return 0, "B", err
	}
	memInfo := string(memInfoBytes)

	return getMemory(memInfo, totalMemRegex)
}

func GetUsedMem() (int64, string, error) {
	return getUsedMemImpl("/proc/meminfo")
}

func getUsedMemImpl(memInfoPath string) (int64, string, error) {
	memInfoBytes, err := os.ReadFile(memInfoPath)
	if err != nil {
		return 0, "B", err
	}
	memInfo := string(memInfoBytes)

	totalMem, unit, err := getMemory(memInfo, totalMemRegex)
	if err != nil {
		return 0, "B", err
	}

	freeMem, _, err := getMemory(memInfo, freeMemRegex)
	if err != nil {
		return 0, "B", err
	}

	return totalMem - freeMem, unit, nil
}

func getMemory(memInfo string, regex *regexp.Regexp) (int64, string, error) {
	if match := regex.FindStringSubmatch(memInfo); match != nil {
		totalMemory, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return 0, "B", err
		}
		units := match[2]
		return totalMemory, units, nil
	} else {
		return 0, "B", nil
	}
}

func GetMemAvail() (int64, error) {
	return getMemAvailImpl("/proc/meminfo")
}

func getMemAvailImpl(memInfoPath string) (int64, error) {
	memInfoBytes, err := os.ReadFile(memInfoPath)
	if err != nil {
		return 0, err
	}

	if match := memAvailRegex.FindStringSubmatch(string(memInfoBytes)); match != nil {
		return strconv.ParseInt(match[1], 10, 64)
	}
	return 0, errors.New("malformed meminfo")
}
