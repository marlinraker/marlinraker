// Package webcams stores webcam configurations, compatible with Moonraker's
// webcam API.
package webcams

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/api/notification"
	"github.com/marlinraker/marlinraker/internal/database"
	"github.com/marlinraker/marlinraker/internal/util"
)

// Webcam is a webcam entry. It is kept as a generic map so fields that newer
// clients send are preserved.
type Webcam map[string]any

// editableFields are the fields a client may set with PostItem.
var editableFields = []string{
	"name", "location", "service", "enabled", "icon", "target_fps", "target_fps_idle",
	"stream_url", "snapshot_url", "flip_horizontal", "flip_vertical", "rotation",
	"aspect_ratio", "extra_data",
}

const (
	dbNamespace = "webcams"
	dbKey       = "items"
)

var mu sync.Mutex

func load() ([]Webcam, error) {
	value, err := database.GetItem(dbNamespace, dbKey, true)
	if err != nil {
		var executorError *util.ExecutorError
		if errors.As(err, &executorError) && executorError.Code == 404 {
			return []Webcam{}, nil
		}
		return nil, err
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	webcams := []Webcam{}
	if err := json.Unmarshal(bytes, &webcams); err != nil {
		return nil, fmt.Errorf("failed to load webcams: %w", err)
	}
	return webcams, nil
}

func save(webcams []Webcam) error {
	if _, err := database.PostItem(dbNamespace, dbKey, webcams, true); err != nil {
		return fmt.Errorf("failed to save webcams: %w", err)
	}
	if err := notification.Publish(notification.New("notify_webcams_changed", []any{map[string]any{"webcams": webcams}})); err != nil {
		log.Errorf("Failed to publish webcam change: %v", err)
	}
	return nil
}

func newUid() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func find(webcams []Webcam, uid string, name string) int {
	return slices.IndexFunc(webcams, func(webcam Webcam) bool {
		if uid != "" {
			return webcam["uid"] == uid
		}
		return webcam["name"] == name
	})
}

// List returns all webcams.
func List() ([]Webcam, error) {
	mu.Lock()
	defer mu.Unlock()
	return load()
}

// GetItem returns the webcam with the given uid or name.
func GetItem(uid string, name string) (Webcam, error) {
	mu.Lock()
	defer mu.Unlock()

	if uid == "" && name == "" {
		return nil, util.NewError(400, "name or uid param is required")
	}
	webcams, err := load()
	if err != nil {
		return nil, err
	}
	idx := find(webcams, uid, name)
	if idx == -1 {
		return nil, util.NewError(404, "webcam not found")
	}
	return webcams[idx], nil
}

// PostItem updates the webcam with the given uid or, without uid, the webcam
// with the given name. A webcam is created if none matches.
func PostItem(params map[string]any) (Webcam, error) {
	mu.Lock()
	defer mu.Unlock()

	webcams, err := load()
	if err != nil {
		return nil, err
	}

	uid, _ := params["uid"].(string)
	name, _ := params["name"].(string)
	if uid == "" && name == "" {
		return nil, util.NewError(400, "name or uid param is required")
	}

	idx := find(webcams, uid, name)
	if idx == -1 {
		if uid != "" {
			return nil, util.NewErrorf(404, "webcam with uid %q not found", uid)
		}
		webcams = append(webcams, Webcam{
			"name":            name,
			"location":        "printer",
			"service":         "mjpegstreamer-adaptive",
			"enabled":         true,
			"icon":            "mdiWebcam",
			"target_fps":      15,
			"target_fps_idle": 5,
			"stream_url":      "",
			"snapshot_url":    "",
			"flip_horizontal": false,
			"flip_vertical":   false,
			"rotation":        0,
			"aspect_ratio":    "4:3",
			"extra_data":      map[string]any{},
			"source":          "database",
			"uid":             newUid(),
		})
		idx = len(webcams) - 1
	}

	if name != "" {
		duplicate := slices.ContainsFunc(webcams, func(other Webcam) bool {
			return other["name"] == name && other["uid"] != webcams[idx]["uid"]
		})
		if duplicate {
			return nil, util.NewErrorf(400, "webcam with name %q already exists", name)
		}
	}

	webcam := webcams[idx]
	for _, field := range editableFields {
		if value, exists := params[field]; exists {
			webcam[field] = value
		}
	}

	return webcam, save(webcams)
}

// DeleteItem deletes the webcam with the given uid or name.
func DeleteItem(uid string, name string) (Webcam, error) {
	mu.Lock()
	defer mu.Unlock()

	if uid == "" && name == "" {
		return nil, util.NewError(400, "name or uid param is required")
	}

	webcams, err := load()
	if err != nil {
		return nil, err
	}

	idx := find(webcams, uid, name)
	if idx == -1 {
		return nil, util.NewError(404, "webcam not found")
	}
	webcam := webcams[idx]
	webcams = slices.Delete(webcams, idx, idx+1)
	return webcam, save(webcams)
}
