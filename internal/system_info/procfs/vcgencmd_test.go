package procfs

import (
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"gotest.tools/assert"
)

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
