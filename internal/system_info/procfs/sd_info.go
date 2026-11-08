package procfs

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SdInfo struct {
	ManufacturerId   string `json:"manufacturer_id,omitempty"`
	Manufacturer     string `json:"manufacturer,omitempty"`
	OemId            string `json:"oem_id,omitempty"`
	ProductName      string `json:"product_name,omitempty"`
	ProductRevision  string `json:"product_revision,omitempty"`
	SerialNumber     string `json:"serial_number,omitempty"`
	ManufacturerDate string `json:"manufacturer_date,omitempty"`
	Capacity         string `json:"capacity,omitempty"`
	TotalBytes       int64  `json:"total_bytes,omitempty"`
}

var manufacturers = map[string]string{
	"1b": "Samsung",
	"03": "Sandisk",
	"74": "PNY",
}

func getSdInfoImpl(cidPath string, csdPath string) (*SdInfo, error) {

	info := &SdInfo{}

	cidBytes, err := os.ReadFile(cidPath)
	if err != nil {
		return info, fmt.Errorf("failed to read cid: %w", err)
	}
	cid := strings.ToLower(strings.TrimSpace(string(cidBytes)))

	csdBytes, err := os.ReadFile(csdPath)
	if err != nil {
		return info, fmt.Errorf("failed to read csd: %w", err)
	}
	csd, err := hex.DecodeString(strings.ToLower(strings.TrimSpace(string(csdBytes))))
	if err != nil {
		return info, fmt.Errorf("failed to decode csd: %w", err)
	}

	info.ManufacturerId = cid[:2]
	info.OemId = cid[2:6]
	info.SerialNumber = cid[18:26]

	var exists bool
	if info.Manufacturer, exists = manufacturers[info.ManufacturerId]; !exists {
		info.Manufacturer = "Unknown"
	}

	if info.ProductName, err = hexToAscii(cid[6:16]); err != nil {
		return nil, fmt.Errorf("failed to parse product name: %w", err)
	}

	if info.ProductRevision, err = parseProductRevision(cid[16:18]); err != nil {
		return nil, fmt.Errorf("failed to parse product revision: %w", err)
	}

	if info.ManufacturerDate, err = parseManufacturerDate(cid[27:30]); err != nil {
		return nil, fmt.Errorf("failed to parse manufacturer date: %w", err)
	}

	switch csd[0] >> 6 & 0x3 {
	case 0:
		maxBlockLenSqrt := int64(csd[5] & 0xf)
		maxBlockLen := maxBlockLenSqrt * maxBlockLenSqrt
		cSize := (int64(csd[6])&0x3)<<10 | int64(csd[7])<<2 | int64(csd[8])>>6&0x3
		cMultReg := int64(csd[9]&0x3<<1 | csd[10]>>7)
		cMult := (cMultReg + 2) * (cMultReg + 2)
		info.TotalBytes = (cSize + 1) * cMult * maxBlockLen
		info.Capacity = fmt.Sprintf("%.1f MiB", float64(info.TotalBytes)/1048576.0)

	case 1:
		cSize := int64(csd[7]&0x3f)<<16 | int64(csd[8])<<8 | int64(csd[9])
		info.TotalBytes = (cSize + 1) * 512 * 1024
		info.Capacity = fmt.Sprintf("%.1f GiB", float64(info.TotalBytes)/1073741824.0)

	case 2:
		cSize := int64(csd[6]&0xf)<<24 | int64(csd[7])<<16 | int64(csd[8])<<8 | int64(csd[9])
		info.TotalBytes = (cSize + 1) * 512 * 1024
		info.Capacity = fmt.Sprintf("%.1f TiB", float64(info.TotalBytes)/1099511627776.0)
	}

	return info, nil
}

func hexToAscii(encoded string) (string, error) {
	bytes, err := hex.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex string: %w", err)
	}
	return string(bytes), nil
}

func parseProductRevision(encoded string) (string, error) {
	major, err := strconv.ParseInt(encoded[:1], 16, 0)
	if err != nil {
		return "", fmt.Errorf("failed to parse product revision: %w", err)
	}
	minor, err := strconv.ParseInt(encoded[1:2], 16, 0)
	if err != nil {
		return "", fmt.Errorf("failed to parse product revision: %w", err)
	}
	return fmt.Sprintf("%d.%d", major, minor), nil
}

func parseManufacturerDate(encoded string) (string, error) {
	year, err := strconv.ParseInt(encoded[0:2], 16, 0)
	if err != nil {
		return "", fmt.Errorf("failed to parse manufacturer date year: %w", err)
	}
	month, err := strconv.ParseInt(encoded[2:3], 16, 0)
	if err != nil {
		return "", fmt.Errorf("failed to parse manufacturer date month: %w", err)
	}
	return fmt.Sprintf("%d/%02d", int(year)+2000, month), nil
}
