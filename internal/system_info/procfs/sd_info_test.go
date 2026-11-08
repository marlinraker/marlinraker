package procfs

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSdInfo(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		info, err := getSdInfoImpl("testdata/cid1", "testdata/csd1")
		assert.NilError(t, err)
		assert.DeepEqual(t, info, &SdInfo{
			ManufacturerId:   "74",
			Manufacturer:     "PNY",
			OemId:            "4a60",
			ProductName:      "USDU1",
			ProductRevision:  "2.0",
			SerialNumber:     "40510971",
			ManufacturerDate: "2018/12",
			Capacity:         "29.4 GiB",
			TotalBytes:       31609323520,
		})
	})

	t.Run("2", func(t *testing.T) {
		info, err := getSdInfoImpl("testdata/cid2", "testdata/csd2")
		assert.NilError(t, err)
		assert.DeepEqual(t, info, &SdInfo{
			ManufacturerId:   "03",
			Manufacturer:     "Sandisk",
			OemId:            "5344",
			ProductName:      "ACLCF",
			ProductRevision:  "8.0",
			SerialNumber:     "e3572a1d",
			ManufacturerDate: "2017/02",
			Capacity:         "119.1 GiB",
			TotalBytes:       127865454592,
		})
	})
}
