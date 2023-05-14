package system_info

import (
	"errors"
	"marlinraker/src/printer_objects"
	"marlinraker/src/util"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type systemStatsObject struct{}

func (systemStatsObject) Query() printer_objects.QueryResult {
	loadAvg, err := getLoadAvg()
	if err != nil {
		util.LogError(err)
		return nil
	}

	processTime, err := getProcessTime()
	if err != nil {
		util.LogError(err)
		return nil
	}

	memAvail, err := getMemAvail()
	if err != nil {
		util.LogError(err)
		return nil
	}

	return printer_objects.QueryResult{
		"sysload":  loadAvg[0],
		"cputime":  processTime,
		"memavail": memAvail,
	}
}

var memAvailRegex = regexp.MustCompile(`(?m)^MemAvailable:\s*([0-9]+) kB$`)

func getLoadAvg() ([3]float32, error) {
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

func getProcessTime() (float32, error) {
	clkBytes, err := exec.Command("getconf", "CLK_TCK").Output()
	if err != nil {
		return 0, err
	}
	clk, err := strconv.ParseInt(strings.TrimSpace(string(clkBytes)), 10, 32)

	pid := os.Getpid()
	return getProcessTimeImpl("/proc/"+strconv.Itoa(pid)+"/stat", int32(clk))
}

func getProcessTimeImpl(procPidStatPath string, clk int32) (float32, error) {
	statBytes, err := os.ReadFile(procPidStatPath)
	if err != nil {
		return 0, err
	}

	parts := strings.Split(strings.TrimSpace(string(statBytes)), " ")
	if len(parts) < 15 {
		return 0, errors.New("malformed " + procPidStatPath)
	}

	utime, err := strconv.ParseInt(parts[13], 10, 64)
	if err != nil {
		return 0, err
	}

	stime, err := strconv.ParseInt(parts[14], 10, 64)
	if err != nil {
		return 0, err
	}

	return float32(utime+stime) / float32(clk), nil
}

func getMemAvail() (int64, error) {
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
