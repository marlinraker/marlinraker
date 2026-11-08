package files

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"
	"gotest.tools/assert"
)

func BenchmarkScanMetadata(b *testing.B) {
	var err error
	DataDir, err = filepath.Abs("testdata/")
	assert.NilError(b, err)
	for i := 0; i < b.N; i++ {
		_, err := ScanMetadata("benchy_prusaslicer.gcode")
		assert.NilError(b, err)
	}
}

func TestMetadata(t *testing.T) {

	Fs = afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())
	const fileName = "benchy_prusaslicer.gcode"

	var err error
	DataDir, err = filepath.Abs("testdata/")
	assert.NilError(t, err)

	metadata, err := ScanMetadata(fileName)
	assert.NilError(t, err)

	// The 32x32 and 300x300 thumbnails are re-encoded by the PNG encoder, whose
	// output size changes between Go versions. Check that the reported sizes match
	// the written files, then ignore them in the comparison below.
	assert.Equal(t, len(metadata.Thumbnails), 3)
	for _, thumbnail := range metadata.Thumbnails {
		stat, err := Fs.Stat(filepath.Join(DataDir, "gcodes", thumbnail.RelativePath))
		assert.NilError(t, err)
		assert.Equal(t, int64(thumbnail.Size), stat.Size())
	}

	assert.DeepEqual(t, metadata, &Metadata{
		FileName:            fileName,
		Size:                4027472,
		Slicer:              "PrusaSlicer",
		SlicerVersion:       "2.5.2",
		LayerHeight:         0.2,
		FirstLayerHeight:    0.2,
		ObjectHeight:        48,
		FilamentTotal:       4198.59,
		EstimatedTime:       5122,
		FirstLayerBedTemp:   60,
		FirstLayerExtrTemp:  215,
		GcodeStartByte:      22773,
		GcodeEndByte:        4015882,
		NozzleDiameter:      0.4,
		FilamentName:        `"Prusament PLA"`,
		FilamentType:        "PLA",
		FilamentWeightTotal: 12.52,
		Thumbnails: []Thumbnail{
			{
				Width:        160,
				Height:       120,
				Size:         16188,
				RelativePath: ".thumbs/benchy_prusaslicer-160x120.png",
			},
			{
				Width:        32,
				Height:       32,
				RelativePath: ".thumbs/benchy_prusaslicer-32x32.png",
			},
			{
				Width:        300,
				Height:       300,
				RelativePath: ".thumbs/benchy_prusaslicer-300x300.png",
			},
		},
	}, cmpopts.IgnoreFields(Metadata{}, "Modified"), cmpopts.IgnoreFields(Thumbnail{}, "Size"))

	err = StoreMetadata(metadata)
	assert.NilError(t, err)

	hasMetadata := HasMetadata(fileName)
	assert.Equal(t, hasMetadata, true)

	read, err := LoadMetadata(fileName)
	assert.NilError(t, err)
	assert.DeepEqual(t, metadata, read)

	err = RemoveMetadata(fileName)
	assert.NilError(t, err)

	hasMetadata = HasMetadata(fileName)
	assert.Equal(t, hasMetadata, false)
}
