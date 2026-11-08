//go:build linux

package procfs

import (
	"os/exec"
)

func GetThrottledState() (ThrottledState, error) {
	throttledBytes, err := exec.Command("vcgencmd", "get_throttled").Output()
	if err != nil {
		return ThrottledState{0, []string{}}, err
	}
	return getThrottledStateImpl(throttledBytes)
}
