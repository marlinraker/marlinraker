package procfs

import (
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/assert"
)

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
