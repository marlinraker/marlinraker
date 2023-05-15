package procfs

import (
	"os/exec"
	"strconv"
	"strings"
)

type ThrottledState struct {
	Bits  int32    `json:"bits"`
	Flags []string `json:"flags"`
}

var bitFlags = map[int]string{
	0:  "Under-Voltage detected",
	1:  "Frequency Capped",
	2:  "Currently Throttled",
	3:  "Temperature Limit Active",
	16: "Previously Under-Volted",
	17: "Previously Frequency Capped",
	18: "Previously Throttled",
	19: "Previously Temperature Limited",
}

func GetThrottledState() (ThrottledState, error) {
	throttledBytes, err := exec.Command("vcgencmd", "get_throttled").Output()
	if err != nil {
		return ThrottledState{0, []string{}}, err
	}
	return getThrottledStateImpl(throttledBytes)
}

func getThrottledStateImpl(throttledBytes []byte) (ThrottledState, error) {
	throttledStr := strings.TrimSpace(string(throttledBytes))[12:]
	bits, err := strconv.ParseInt(throttledStr, 16, 32)
	if err != nil {
		return ThrottledState{0, []string{}}, err
	}

	flags := make([]string, 0)
	for bit, flag := range bitFlags {
		if bits&(1<<bit) != 0 {
			flags = append(flags, flag)
		}
	}

	return ThrottledState{Bits: int32(bits), Flags: flags}, nil
}
