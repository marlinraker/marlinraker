package procfs

import (
	"testing"

	"gotest.tools/v3/assert"
)

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
