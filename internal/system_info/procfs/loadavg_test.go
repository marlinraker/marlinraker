package procfs

import (
	"testing"

	"gotest.tools/assert"
)

func TestLoadAvg(t *testing.T) {
	avg, err := getLoadAvgImpl("testdata/loadavg")
	assert.NilError(t, err)
	assert.DeepEqual(t, avg, [3]float32{2.02, 2.33, 2.40})
}
