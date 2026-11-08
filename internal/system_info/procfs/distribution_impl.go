//go:build linux

package procfs

func GetDistribution() (*Distribution, error) {
	return getDistributionImpl("/etc/os-release")
}
