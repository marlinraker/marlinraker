package api

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"marlinraker/src/api/executors"
	"marlinraker/src/marlinraker"
	"marlinraker/src/util"
	"net/http"
	"strings"
)

var (
	apiVersionResult = map[string]any{
		"server": "1.5.0",
		"api":    "0.1",
		"text":   "OctoPrint (Marlinraker)",
	}

	apiServerResult = map[string]any{
		"server":   "1.5.0",
		"safemode": "settings",
	}

	apiLoginResult = map[string]any{
		"_is_external_client": false,
		"_login_mechanism":    "apikey",
		"name":                "_api",
		"active":              true,
		"user":                true,
		"admin":               true,
		"apikey":              nil,
		"permissions":         []string{},
		"groups":              []string{"admins", "users"},
	}

	apiSettingsResult = map[string]any{
		"plugins": map[string]any{
			"UltimakerFormatPackage": map[string]any{
				"align_inline_thumbnail":       false,
				"inline_thumbnail":             false,
				"inline_thumbnail_align_value": "left",
				"inline_thumbnail_scale_value": "50",
				"installed":                    true,
				"installed_version":            "0.2.2",
				"scale_inline_thumbnail":       false,
				"state_panel_thumbnail":        true,
			},
		},
		"feature": map[string]any{
			"sdSupport":        false,
			"temperatureGraph": false,
		},
		"webcam": map[string]any{
			"flipH":         false,
			"flipV":         false,
			"rotate90":      false,
			"streamUrl":     "/webcam/?action=stream",
			"webcamEnabled": true,
		},
	}

	apiJobResult = map[string]any{
		"job": map[string]any{
			"file": map[string]any{
				"name": nil,
			},
			"estimatedPrintTime": nil,
			"filament": map[string]any{
				"length": nil,
			},
			"user": nil,
		},
		"progress": map[string]any{
			"completion":      nil,
			"filepos":         nil,
			"printTime":       nil,
			"printTimeLeft":   nil,
			"printTimeOrigin": nil,
		},
		"state": "Offline",
	}

	apiPrinterResult = map[string]any{
		"temperature": map[string]any{
			"tool0": map[string]any{
				"actual": 22.25,
				"offset": 0.,
				"target": 0.,
			},
			"bed": map[string]any{
				"actual": 22.25,
				"offset": 0.,
				"target": 0.,
			},
		},
		"state": map[string]any{
			"text": "state",
			"flags": map[string]any{
				"operational":   true,
				"paused":        false,
				"printing":      false,
				"cancelling":    false,
				"pausing":       false,
				"error":         false,
				"ready":         false,
				"closedOrError": false,
			},
		},
	}

	apiPrinterProfilesResult = map[string]any{
		"profiles": map[string]any{
			"_default": map[string]any{
				"id":            "_default",
				"name":          "Default",
				"color":         "default",
				"model":         "Default",
				"default":       true,
				"current":       true,
				"heatedBed":     true,
				"heatedChamber": false,
			},
		},
	}
)

func handleOctoPrint(writer http.ResponseWriter, request *http.Request) error {

	method := request.Method
	path := strings.TrimRight(request.URL.Path, "/")

	var result any

	if method == "GET" {
		switch path {
		case "/api/version":
			result = apiVersionResult
		case "/api/server":
			result = apiServerResult
		case "/api/login":
			result = apiLoginResult
		case "/api/settings":
			result = apiSettingsResult
		case "/api/job":
			result = apiJobResult
		case "/api/printer":
			result = apiPrinterResult
		case "/api/printerprofiles":
			result = apiPrinterProfilesResult
		}

	} else if method == "POST" {
		switch path {
		case "/api/files/local":
			var err error
			result, err = executors.ServerFilesUpload(nil, request, nil)
			if err != nil {
				return err
			}
		case "/api/printer/command":
			result = handleApiPrinterCommand(request)
		}
	}

	if result == nil {
		log.Errorln("Cannot find OctoPrint API endpoint \"" + method + " " + path + "\"")
		writer.WriteHeader(404)
		_, err := writer.Write([]byte("Not found"))
		return err
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	writer.WriteHeader(200)
	_, err = writer.Write(bytes)
	return err
}

func handleApiPrinterCommand(request *http.Request) any {

	printer := marlinraker.Printer
	if printer == nil {
		return struct{}{}
	}

	if request.Body != nil && request.ContentLength > 0 {
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			return err
		}
		var body struct {
			Commands []string `json:"commands"`
		}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			util.LogError(err)
			return struct{}{}
		}

		for _, command := range body.Commands {
			<-printer.MainExecutorContext().QueueGcode(command, false)
		}
	}
	return struct{}{}
}
