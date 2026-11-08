//go:build linux

package procfs

func GetSdInfo() (*SdInfo, error) {
	return getSdInfoImpl("/sys/block/mmcblk0/device/cid", "/sys/block/mmcblk0/device/csd")
}
