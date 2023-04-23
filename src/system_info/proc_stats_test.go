package system_info

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"gotest.tools/assert"
	"testing"
)

func BenchmarkTakeSnapshot(b *testing.B) {
	err := takeSnapshot()
	if err != nil {
		b.Fatal(err)
	}
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

		usage := getCpuUsage(last, curr)

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

		usage := getCpuUsage(last, curr)

		assert.DeepEqual(t, usage, map[string]float64{
			"cpu":  100.0,
			"cpu0": 100.0,
		}, cmpopts.EquateApprox(0.0, 1e-9))
	})
}

func TestCpuTemp(t *testing.T) {
	_, err := getCpuTemp()
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

	assert.DeepEqual(t, second.Stats, networkStats{
		"br-0e140c4dcebf": {RxBytes: 79955, TxBytes: 671178},
		"docker0":         {RxBytes: 20844, TxBytes: 412148},
		"enp0s31f6":       {RxBytes: 1640916711, TxBytes: 50845348, Bandwidth: 117300},
		"lo":              {RxBytes: 2148053, TxBytes: 2148053},
		"veth14520d9":     {RxBytes: 21292, TxBytes: 447867},
		"veth8103a4a":     {RxBytes: 87627, TxBytes: 706167},
	})
}

func TestUptime(t *testing.T) {
	_, err := getUptime()
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
		assert.DeepEqual(t, state, throttledState{
			Bits:  0,
			Flags: []string{},
		})
	})

	t.Run("2", func(t *testing.T) {
		state, err := getThrottledStateImpl([]byte("throttled=0xF000F\n"))
		if err != nil {
			t.Fatal(err)
		}
		assert.DeepEqual(t, state, throttledState{
			Bits:  0xF000F,
			Flags: lo.Values(bitFlags),
		}, cmpopts.SortSlices(func(a string, b string) bool { return a < b }))
	})
}
