//go:build linux

package procfs

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func GetProcessTime() (float32, error) {
	clkBytes, err := exec.Command("getconf", "CLK_TCK").Output()
	if err != nil {
		return 0, err
	}
	clk, err := strconv.ParseInt(strings.TrimSpace(string(clkBytes)), 10, 32)
	if err != nil {
		return 0, err
	}
	if clk <= 0 {
		return 0, fmt.Errorf("invalid CLK_TCK value %d", clk)
	}

	pid := os.Getpid()
	return getProcessTimeImpl(fmt.Sprintf("/proc/%d/stat", pid), int32(clk))
}
