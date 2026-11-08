//go:build linux

package procfs

import "time"

func GetNetworkStats(lastStats *TimedNetworkStats) (*TimedNetworkStats, error) {
	now := float64(time.Now().UnixMilli()) / 1000.0
	return getNetworkStatsImpl(lastStats, now, "/proc/net/dev")
}
