package system_info

import (
	"gotest.tools/assert"
	"testing"
)

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
