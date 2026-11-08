package gcode_store

import (
	"time"
)

type GcodeResponseType string

const (
	Command  GcodeResponseType = "command"
	Response GcodeResponseType = "response"
)

type GcodeLog struct {
	Message string            `json:"message"`
	Time    float64           `json:"time"`
	Type    GcodeResponseType `json:"type"`
}

var GcodeStore = make([]GcodeLog, 0)

func LogNow(message string, responseType GcodeResponseType) {
	now := float64(time.Now().UnixMilli()) / 1000.0
	Log(message, now, responseType)
}

func Log(message string, time float64, responseType GcodeResponseType) {
	GcodeStore = append(GcodeStore, GcodeLog{message, time, responseType})
	if len(GcodeStore) > 1000 {
		GcodeStore = GcodeStore[1:]
	}
}
