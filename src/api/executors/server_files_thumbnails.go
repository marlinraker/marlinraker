package executors

import (
	"github.com/samber/lo"
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"net/http"
	"path/filepath"
)

type ServerFilesThumbnailsResult *files.Metadata

type thumbnail struct {
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Size          int    `json:"size"`
	ThumbnailPath string `json:"thumbnail_path"`
}

func ServerFilesThumbnails(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	fileName, err := params.RequireString("filename")
	if err != nil {
		return nil, err
	}

	metadata, err := files.LoadOrScanMetadata(fileName)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(fileName)
	return lo.Map(metadata.Thumbnails, func(thumb files.Thumbnail, _ int) thumbnail {
		return thumbnail{
			Width:         thumb.Width,
			Height:        thumb.Height,
			Size:          thumb.Size,
			ThumbnailPath: filepath.Join(dir, thumb.RelativePath),
		}
	}), nil
}
