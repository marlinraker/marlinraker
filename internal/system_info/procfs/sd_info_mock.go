//go:build !linux

package procfs

func GetSdInfo() (*SdInfo, error) {
	return &SdInfo{
		ManufacturerId:   "",
		Manufacturer:     "Unknown",
		OemId:            "",
		ProductName:      "Unknown",
		ProductRevision:  "",
		SerialNumber:     "",
		ManufacturerDate: "",
		Capacity:         "",
		TotalBytes:       0,
	}, nil
}
