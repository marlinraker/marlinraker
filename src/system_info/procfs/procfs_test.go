package procfs

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"gotest.tools/assert"
	"testing"
)

func TestCpuUsage(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		last, err := getCpuTimesImpl("testdata/stat1_1")
		if err != nil {
			t.Fatal(err)
		}

		curr, err := getCpuTimesImpl("testdata/stat1_2")
		if err != nil {
			t.Fatal(err)
		}

		usage := GetCpuUsage(last, curr)

		assert.DeepEqual(t, usage, map[string]float64{
			"cpu":  17.0,
			"cpu0": 5.2,
			"cpu1": 16.4,
			"cpu2": 98.8,
			"cpu3": 6.6,
			"cpu4": 1.3,
			"cpu5": 2.7,
			"cpu6": 2.8,
			"cpu7": 2.2,
		}, cmpopts.EquateApprox(0.0, 1e-9))
	})

	t.Run("2", func(t *testing.T) {
		last, err := getCpuTimesImpl("testdata/stat2_1")
		if err != nil {
			t.Fatal(err)
		}

		curr, err := getCpuTimesImpl("testdata/stat2_2")
		if err != nil {
			t.Fatal(err)
		}

		usage := GetCpuUsage(last, curr)

		assert.DeepEqual(t, usage, map[string]float64{
			"cpu":  100.0,
			"cpu0": 100.0,
		}, cmpopts.EquateApprox(0.0, 1e-9))
	})
}

func TestCpuTemp(t *testing.T) {
	_, err := getCpuTempImpl("testdata/temp")
	if err != nil {
		t.Fatal(err)
	}
}

func TestNetworkStats(t *testing.T) {

	first, err := getNetworkStatsImpl(nil, 0.0, "testdata/net_1")
	if err != nil {
		t.Fatal(err)
	}

	second, err := getNetworkStatsImpl(first, 1.0, "testdata/net_2")
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, second.Stats, NetworkStats{
		"br-0e140c4dcebf": {RxBytes: 79955, TxBytes: 671178},
		"docker0":         {RxBytes: 20844, TxBytes: 412148},
		"enp0s31f6":       {RxBytes: 1640916711, TxBytes: 50845348, Bandwidth: 117300},
		"lo":              {RxBytes: 2148053, TxBytes: 2148053},
		"veth14520d9":     {RxBytes: 21292, TxBytes: 447867},
		"veth8103a4a":     {RxBytes: 87627, TxBytes: 706167},
	})
}

func TestUptime(t *testing.T) {
	_, err := GetUptime()
	if err != nil {
		t.Fatal(err)
	}
}

func TestThrottledState(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		state, err := getThrottledStateImpl([]byte("throttled=0x0\n"))
		if err != nil {
			t.Fatal(err)
		}
		assert.DeepEqual(t, state, ThrottledState{
			Bits:  0,
			Flags: []string{},
		})
	})

	t.Run("2", func(t *testing.T) {
		state, err := getThrottledStateImpl([]byte("throttled=0xF000F\n"))
		if err != nil {
			t.Fatal(err)
		}
		assert.DeepEqual(t, state, ThrottledState{
			Bits:  0xF000F,
			Flags: lo.Values(bitFlags),
		}, cmpopts.SortSlices(func(a string, b string) bool { return a < b }))
	})
}

func TestCpuInfo(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		info, err := getCpuInfoImpl("testdata/cpuinfo1", "testdata/meminfo1")
		assert.NilError(t, err)
		assert.Check(t, info.Bits == "32bit" || info.Bits == "64bit", info.Bits)
		assert.Check(t, info.Processor != "", info.Processor)
		assert.DeepEqual(t, info, &CpuInfo{
			CpuCount:     8,
			CpuDesc:      "Intel(R) Core(TM) i7-6700K CPU @ 4.00GHz",
			SerialNumber: "",
			HardwareDesc: "",
			Model:        "",
			TotalMemory:  32805556,
			MemoryUnits:  "kB",
		}, cmpopts.IgnoreFields(CpuInfo{}, "Processor", "Bits"))
	})

	t.Run("2", func(t *testing.T) {
		info, err := getCpuInfoImpl("testdata/cpuinfo2", "testdata/meminfo2")
		assert.NilError(t, err)
		assert.Check(t, info.Bits == "32bit" || info.Bits == "64bit", info.Bits)
		assert.Check(t, info.Processor != "", info.Processor)
		assert.DeepEqual(t, info, &CpuInfo{
			CpuCount:     1,
			CpuDesc:      "ARMv6-compatible processor rev 7 (v6l)",
			SerialNumber: "00000000ed053e32",
			HardwareDesc: "BCM2835",
			Model:        "Raspberry Pi Zero W Rev 1.1",
			TotalMemory:  439592,
			MemoryUnits:  "kB",
		}, cmpopts.IgnoreFields(CpuInfo{}, "Processor", "Bits"))
	})
}

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

func TestDistribution(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		info, err := getDistributionImpl("testdata/os-release1")
		assert.NilError(t, err)
		assert.DeepEqual(t, info, &Distribution{
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
		assert.DeepEqual(t, info, &Distribution{
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
	_, err := GetNetwork()
	assert.NilError(t, err)
}

func TestLoadAvg(t *testing.T) {
	avg, err := getLoadAvgImpl("testdata/loadavg")
	assert.NilError(t, err)
	assert.DeepEqual(t, avg, [3]float32{2.02, 2.33, 2.40})
}

func TestProcessTime(t *testing.T) {
	time, err := getProcessTimeImpl("testdata/proc_pid_stat", 100)
	assert.NilError(t, err)
	assert.Equal(t, time, float32(0.02))
}

func TestMemAvail(t *testing.T) {
	avail, err := getMemAvailImpl("testdata/meminfo1")
	assert.NilError(t, err)
	assert.Equal(t, avail, int64(24552156))

	avail, err = getMemAvailImpl("testdata/meminfo2")
	assert.NilError(t, err)
	assert.Equal(t, avail, int64(302256))
}
