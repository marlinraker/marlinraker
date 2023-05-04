package system_info

import (
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type cpuInfo struct {
	CpuCount     int    `json:"cpu_count"`
	Bits         string `json:"bits"`
	Processor    string `json:"processor"`
	CpuDesc      string `json:"cpu_desc"`
	SerialNumber string `json:"serial_number"`
	HardwareDesc string `json:"hardware_desc"`
	Model        string `json:"model"`
	TotalMemory  int64  `json:"total_memory"`
	MemoryUnits  string `json:"memory_units"`
}

var (
	serialRegex    = regexp.MustCompile(`(?m)^Serial\s*: (.+)$`)
	hardwareRegex  = regexp.MustCompile(`(?m)^Hardware\s*: (.+)$`)
	modelRegex     = regexp.MustCompile(`(?m)^Model\s*: (.+)$`)
	modelNameRegex = regexp.MustCompile(`(?m)^model name\s*: (.+)$`)
	processorRegex = regexp.MustCompile(`(?m)^processor\s*: [0-9]+$`)
)

func getCpuInfo() (*cpuInfo, error) {
	return getCpuInfoImpl("/proc/cpuinfo", "/proc/meminfo")
}

func getCpuInfoImpl(cpuInfoPath string, memInfoPath string) (*cpuInfo, error) {

	info := &cpuInfo{}

	cpuInfoBytes, err := os.ReadFile(cpuInfoPath)
	if err != nil {
		return info, err
	}
	cpuInfo := string(cpuInfoBytes)

	if matches := processorRegex.FindAllString(cpuInfo, -1); matches != nil {
		info.CpuCount = len(matches)
	} else {
		info.CpuCount = 1
	}

	if out, err := exec.Command("getconf", "LONG_BIT").Output(); err == nil {
		info.Bits = strings.TrimSpace(string(out)) + "bit"
	} else {
		return info, err
	}

	if out, err := exec.Command("uname", "-m").Output(); err == nil {
		info.Processor = strings.TrimSpace(string(out))
	} else {
		return info, err
	}

	if match := modelNameRegex.FindStringSubmatch(cpuInfo); match != nil {
		info.CpuDesc = match[1]
	}

	if match := serialRegex.FindStringSubmatch(cpuInfo); match != nil {
		info.SerialNumber = match[1]
	}

	if match := hardwareRegex.FindStringSubmatch(cpuInfo); match != nil {
		info.HardwareDesc = match[1]
	}

	if match := modelRegex.FindStringSubmatch(cpuInfo); match != nil {
		info.Model = match[1]
	}

	if info.TotalMemory, info.MemoryUnits, err = getTotalMemImpl(memInfoPath); err != nil {
		info.TotalMemory = 0
		info.MemoryUnits = "B"
	}

	return info, nil
}

type cpuTimes = map[string][]int64

var cpuRegex = regexp.MustCompile(`(?m)^(cpu[0-9]*)\s*(.*)$`)

func getCpuTimes() (cpuTimes, error) {
	return getCpuTimesImpl("/proc/stat")
}

func getCpuTimesImpl(procStatPath string) (cpuTimes, error) {

	times := make(cpuTimes)

	procStatBytes, err := os.ReadFile(procStatPath)
	if err != nil {
		return times, err
	}
	procStats := string(procStatBytes)

	if matches := cpuRegex.FindAllStringSubmatch(procStats, -1); matches != nil {

		for _, match := range matches {
			cpu := match[1]
			parts := strings.Split(match[2], " ")
			_times := make([]int64, len(parts))

			for index, part := range parts {
				time, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					return nil, err
				}
				_times[index] = time
			}

			times[cpu] = _times
		}
	}

	return times, nil
}

type cpuUsage = map[string]float64

func getCpuUsage(lastTimes cpuTimes, currentTimes cpuTimes) cpuUsage {

	usage := make(cpuUsage)

	for cpu, current := range currentTimes {
		var (
			last       = lastTimes[cpu]
			sum  int64 = 0
			idle int64 = 0
		)
		for i := 0; i < len(current); i++ {
			sum += current[i] - last[i]
			if i == 3 {
				idle = current[i] - last[i]
			}
		}
		usage[cpu] = 100.0 - math.Floor(float64(idle)/float64(sum)*1000.0)/10.0
	}

	return usage
}

func getCpuTemp() (float64, error) {
	return getCpuTempImpl("/sys/class/thermal/thermal_zone0/temp")
}

func getCpuTempImpl(thermalZonePath string) (float64, error) {

	tempBytes, err := os.ReadFile(thermalZonePath)
	if err != nil {
		return 0, err
	}

	temp, err := strconv.ParseInt(strings.TrimSpace(string(tempBytes)), 10, 0)
	if err != nil {
		return 0, err
	}

	return math.Round(float64(temp)/10.0) / 100.0, nil
}
