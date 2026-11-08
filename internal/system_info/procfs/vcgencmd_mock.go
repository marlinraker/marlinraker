//go:build !linux

package procfs

func GetThrottledState() (ThrottledState, error) {
	return ThrottledState{0, []string{}}, nil
}
