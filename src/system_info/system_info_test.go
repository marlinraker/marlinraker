package system_info

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/assert"
	"testing"
)

func BenchmarkGetSystemInfo(b *testing.B) {
	_, err := GetSystemInfo()
	assert.NilError(b, err)
}

func TestCpuInfo(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		info, err := getCpuInfoImpl("testdata/cpuinfo1", "testdata/meminfo1")
		assert.NilError(t, err)
		assert.Check(t, info.Bits == "32bit" || info.Bits == "64bit", info.Bits)
		assert.Check(t, info.Processor != "", info.Processor)
		assert.DeepEqual(t, info, &cpuInfo{
			CpuCount:     8,
			CpuDesc:      "Intel(R) Core(TM) i7-6700K CPU @ 4.00GHz",
			SerialNumber: "",
			HardwareDesc: "",
			Model:        "",
			TotalMemory:  32805556,
			MemoryUnits:  "kB",
		}, cmpopts.IgnoreFields(cpuInfo{}, "Processor", "Bits"))
	})

	t.Run("2", func(t *testing.T) {
		info, err := getCpuInfoImpl("testdata/cpuinfo2", "testdata/meminfo2")
		assert.NilError(t, err)
		assert.Check(t, info.Bits == "32bit" || info.Bits == "64bit", info.Bits)
		assert.Check(t, info.Processor != "", info.Processor)
		assert.DeepEqual(t, info, &cpuInfo{
			CpuCount:     1,
			CpuDesc:      "ARMv6-compatible processor rev 7 (v6l)",
			SerialNumber: "00000000ed053e32",
			HardwareDesc: "BCM2835",
			Model:        "Raspberry Pi Zero W Rev 1.1",
			TotalMemory:  439592,
			MemoryUnits:  "kB",
		}, cmpopts.IgnoreFields(cpuInfo{}, "Processor", "Bits"))
	})
}

func TestSdInfo(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		info, err := getSdInfoImpl("testdata/cid1", "testdata/csd1")
		assert.NilError(t, err)
		assert.DeepEqual(t, info, &sdInfo{
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
		assert.DeepEqual(t, info, &sdInfo{
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

func TestDistribution(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		info, err := getDistributionImpl("testdata/os-release1")
		assert.NilError(t, err)
		assert.DeepEqual(t, info, &distribution{
			Name:    "Ubuntu 22.04.2 LTS",
			Id:      "ubuntu",
			Version: "22.04",
			VersionParts: versionParts{
				Major: "22",
				Minor: "04",
			},
			Like:     "debian",
			Codename: "jammy",
		})
	})

	t.Run("2", func(t *testing.T) {
		info, err := getDistributionImpl("testdata/os-release2")
		assert.NilError(t, err)
		assert.DeepEqual(t, info, &distribution{
			Name:    "Raspbian GNU/Linux 11 (bullseye)",
			Id:      "raspbian",
			Version: "11",
			VersionParts: versionParts{
				Major: "11",
			},
			Like:     "debian",
			Codename: "bullseye",
		})
	})
}

func TestNetworks(t *testing.T) {
	_, err := getNetwork()
	assert.NilError(t, err)
}
