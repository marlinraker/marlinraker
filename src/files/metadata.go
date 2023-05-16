package files

import (
	"encoding/gob"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"io/fs"
	"marlinraker/src/util"
	"os"
	"path/filepath"
	"strconv"
)

type Thumbnail struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Size         int    `json:"size"`
	RelativePath string `json:"relative_path"`
}

type Metadata struct {
	FileName            string      `json:"filename"`
	Size                int64       `json:"size"`
	Modified            float64     `json:"modified"`
	PrintStartTime      float64     `json:"print_start_time,omitempty"`
	JobId               string      `json:"job_id,omitempty"`
	Slicer              string      `json:"slicer,omitempty"`
	SlicerVersion       string      `json:"slicer_version,omitempty"`
	LayerHeight         float64     `json:"layer_height,omitempty"`
	FirstLayerHeight    float64     `json:"first_layer_height,omitempty"`
	ObjectHeight        float64     `json:"object_height,omitempty"`
	FilamentTotal       float64     `json:"filament_total,omitempty"`
	EstimatedTime       float64     `json:"estimated_time,omitempty"`
	Thumbnails          []Thumbnail `json:"thumbnails,omitempty"`
	FirstLayerBedTemp   float64     `json:"first_layer_bed_temp,omitempty"`
	FirstLayerExtrTemp  float64     `json:"first_layer_extr_temp,omitempty"`
	GcodeStartByte      int64       `json:"gcode_start_byte,omitempty"`
	GcodeEndByte        int64       `json:"gcode_end_byte,omitempty"`
	NozzleDiameter      float64     `json:"nozzle_diameter,omitempty"`
	FilamentName        string      `json:"filament_name,omitempty"`
	FilamentType        string      `json:"filament_type,omitempty"`
	FilamentWeightTotal float64     `json:"filament_weight_total,omitempty"`
}

func RemoveUnusedMetadata() error {
	gcodeRootPath := filepath.Join(DataDir, "gcodes")
	used := make([]string, 0)

	err := afero.Walk(Fs, gcodeRootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(info.Name())
		if ext == ".meta" || ext == ".png" {
			return nil
		}

		relPath, err := filepath.Rel(gcodeRootPath, path)
		if err != nil {
			return err
		}

		meta, err := LoadMetadata(relPath)
		if err == nil {
			metaDiskPath := getMetaDiskPath(relPath)
			relMetaPath, err := filepath.Rel(gcodeRootPath, metaDiskPath)
			if err != nil {
				return err
			}

			used = append(used, relMetaPath)
			for _, thumbnail := range meta.Thumbnails {
				used = append(used, filepath.Join(filepath.Dir(relPath), thumbnail.RelativePath))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	var deleted int
	err = afero.Walk(Fs, gcodeRootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(gcodeRootPath, path)
		if err != nil {
			return err
		}

		ext, dir := filepath.Ext(path), filepath.Dir(path)
		if ext == ".meta" {
			if !lo.Contains(used, relPath) {
				if err := Fs.Remove(path); err != nil {
					return err
				}
				deleted++
			}
		} else if ext == ".png" && filepath.Base(dir) == ".thumbs" {
			if !lo.Contains(used, relPath) {
				if err := Fs.Remove(path); err != nil {
					return err
				}
				if isEmpty, err := afero.IsEmpty(Fs, dir); err != nil {
					return err
				} else if isEmpty {
					if err := Fs.Remove(dir); err != nil {
						return err
					}
				}
				deleted++
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func HasMetadata(fileName string) bool {
	diskPath := getMetaDiskPath(fileName)
	_, err := Fs.Stat(diskPath)
	return err == nil
}

func LoadOrScanMetadata(fileName string) (*Metadata, error) {

	metadata, err := LoadMetadata(fileName)
	if err == nil {
		log.Debugln("Loaded metadata for \"" + fileName + "\" from disk")
		return metadata, nil
	}

	metadata, err = ScanMetadata(fileName)
	if err != nil {
		return nil, err
	}
	if err := StoreMetadata(metadata); err != nil {
		return nil, err
	}
	log.Debugln("Scanned and stored metadata for \"" + fileName + "\"")
	return metadata, nil
}

func StoreMetadata(metadata *Metadata) error {

	diskPath := getMetaDiskPath(metadata.FileName)

	file, err := Fs.OpenFile(diskPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			util.LogError(err)
		}
	}()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(metadata); err != nil {
		return err
	}
	return nil
}

func LoadMetadata(fileName string) (*Metadata, error) {

	diskPath := getMetaDiskPath(fileName)

	file, err := Fs.Open(diskPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			util.LogError(err)
		}
	}()

	metadata := &Metadata{}
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

func MoveMetadata(src string, dest string) error {

	metadata, err := LoadMetadata(src)
	if err != nil {
		return err
	}

	srcMetaDisk, destMetaDisk := getMetaDiskPath(src), getMetaDiskPath(dest)
	srcDir, destDir := filepath.Dir(srcMetaDisk), filepath.Dir(destMetaDisk)
	thumbDir := filepath.Join(destDir, ".thumbs/")
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return err
	}

	destBaseName := filepath.Base(dest)
	for i, thumbnail := range metadata.Thumbnails {
		destRelPath := ".thumbs/" + destBaseName + "-" + strconv.Itoa(thumbnail.Width) + "x" + strconv.Itoa(thumbnail.Height) + ".png"
		srcDiskPath, destDiskPath := filepath.Join(srcDir, thumbnail.RelativePath), filepath.Join(destDir, destRelPath)
		if err := Fs.Rename(srcDiskPath, destDiskPath); err != nil {
			return err
		}
		thumbnail.RelativePath = destRelPath
		metadata.Thumbnails[i] = thumbnail
	}

	srcThumbDir := filepath.Join(srcDir, ".thumbs/")
	if isEmpty, err := afero.IsEmpty(Fs, srcThumbDir); err != nil {
		return err
	} else if isEmpty {
		if err := Fs.Remove(srcThumbDir); err != nil {
			return err
		}
	}

	metadata.FileName = dest
	if err := StoreMetadata(metadata); err != nil {
		return err
	}

	return Fs.Remove(srcMetaDisk)
}

func DirectoryRenamed(dir string) error {

	gcodeRootPath := filepath.Join(DataDir, "gcodes")
	diskPath := filepath.Join(gcodeRootPath, dir)
	return afero.Walk(Fs, diskPath, func(path string, info fs.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(gcodeRootPath, path)
		if err != nil {
			return err
		}
		metadata, err := LoadMetadata(relPath)
		if err == nil {
			metadata.FileName = relPath
			_ = StoreMetadata(metadata)
		}
		return nil
	})
}

func RemoveMetadata(fileName string) error {

	metadata, err := LoadMetadata(fileName)
	if err != nil {
		return err
	}

	dir := filepath.Join(DataDir, "gcodes", filepath.Dir(fileName))
	for _, thumbnail := range metadata.Thumbnails {
		diskPath := filepath.Join(dir, thumbnail.RelativePath)
		if err := Fs.Remove(diskPath); err != nil {
			return err
		}
	}

	metaDiskPath := getMetaDiskPath(fileName)
	return Fs.Remove(metaDiskPath)
}

func getMetaDiskPath(fileName string) string {
	gcodeDiskPath := filepath.Join(DataDir, "gcodes", fileName)
	dir, base := filepath.Dir(gcodeDiskPath), filepath.Base(gcodeDiskPath)
	metaDiskPath := filepath.Join(dir, "."+base+".meta")
	return metaDiskPath
}
