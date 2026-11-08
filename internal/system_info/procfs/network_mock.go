//go:build !linux

package procfs

func GetNetworkStats(*TimedNetworkStats) (*TimedNetworkStats, error) {
	return &TimedNetworkStats{
		Stats: NetworkStats{},
		Time:  0,
	}, nil
}
