package procfs

import (
	"testing"

	"gotest.tools/assert"
)

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
