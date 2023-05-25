package api

import (
	"bytes"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"marlinraker/src/api/notification"
	"marlinraker/src/config"
	"marlinraker/src/files"
	"marlinraker/src/marlinraker"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
)

func testOctoPrintEndpoint(t *testing.T, method string, endpoint string, body []byte,
	f func(*testing.T, string)) {
	t.Run(method+endpoint, func(t *testing.T) {

		urlQuery, err := url.Parse(endpoint)
		if err != nil {
			t.Fatal(err)
		}

		request, _ := http.NewRequest(method, urlQuery.String(), bytes.NewReader(body))

		recorder := httptest.NewRecorder()
		HttpHandler{}.ServeHTTP(recorder, request)

		f(t, recorder.Body.String())
	})
}

func TestOctoPrint(t *testing.T) {

	marlinraker.Config = config.DefaultConfig()
	marlinraker.Config.Misc.OctoprintCompat = true

	testOctoPrintEndpoint(t, "GET", "/api/version", nil, func(t *testing.T, result string) {
		assert.JSONEq(t, `
		{
            "server": "1.5.0",
            "api": "0.1",
            "text": "OctoPrint (Marlinraker)"
        }
		`, result)
	})

	testOctoPrintEndpoint(t, "GET", "/api/server", nil, func(t *testing.T, result string) {
		assert.JSONEq(t, `
		{
            "server": "1.5.0",
            "safemode": "settings"
        }
		`, result)
	})

	testOctoPrintEndpoint(t, "GET", "/api/login", nil, func(t *testing.T, result string) {
		assert.JSONEq(t, `
		{
            "_is_external_client": false,
            "_login_mechanism": "apikey",
            "name": "_api",
            "active": true,
            "user": true,
            "admin": true,
            "apikey": null,
            "permissions": [],
            "groups": ["admins", "users"]
        }
		`, result)
	})

	testOctoPrintEndpoint(t, "GET", "/api/settings", nil, func(t *testing.T, result string) {
		assert.JSONEq(t, `
		{
            "plugins": {
                "UltimakerFormatPackage": {
                    "align_inline_thumbnail": false,
                    "inline_thumbnail": false,
                    "inline_thumbnail_align_value": "left",
                    "inline_thumbnail_scale_value": "50",
                    "installed": true,
                    "installed_version": "0.2.2",
                    "scale_inline_thumbnail": false,
                    "state_panel_thumbnail": true
                }
            },
            "feature": {
                "sdSupport": false,
                "temperatureGraph": false
            },
            "webcam": {
                "flipH": false,
                "flipV": false,
                "rotate90": false,
                "streamUrl": "/webcam/?action=stream",
                "webcamEnabled": true
            }
        }
		`, result)
	})

	testOctoPrintEndpoint(t, "GET", "/api/job", nil, func(t *testing.T, result string) {
		assert.JSONEq(t, `
		{
            "job": {
                "file": {
                    "name": null
                },
                "estimatedPrintTime": null,
                "filament": {
                    "length": null
                },
                "user": null
            },
            "progress": {
                "completion": null,
                "filepos": null,
                "printTime": null,
                "printTimeLeft": null,
                "printTimeOrigin": null
            },
            "state": "Offline"
        }
		`, result)
	})

	testOctoPrintEndpoint(t, "GET", "/api/printer", nil, func(t *testing.T, result string) {
		assert.JSONEq(t, `
		{
            "temperature": {
                "tool0": {
                    "actual": 22.25,
                    "offset": 0,
                    "target": 0
                },
                "bed": {
                    "actual": 22.25,
                    "offset": 0,
                    "target": 0
                }
            },
            "state": {
                "text": "state",
                "flags": {
                    "operational": true,
                    "paused": false,
                    "printing": false,
                    "cancelling": false,
                    "pausing": false,
                    "error": false,
                    "ready": false,
                    "closedOrError": false
                }
            }
        }
		`, result)
	})

	testOctoPrintEndpoint(t, "GET", "/api/printerprofiles", nil, func(t *testing.T, result string) {
		assert.JSONEq(t, `
		{
            "profiles": {
                "_default": {
                    "id": "_default",
                    "name": "Default",
                    "color": "default",
                    "model": "Default",
                    "default": true,
                    "current": true,
                    "heatedBed": true,
                    "heatedChamber": false
                }
            }
        }
		`, result)
	})

	testOctoPrintEndpoint(t, "POST", "/api/printer/command", []byte(`{"commands":["M115","G28"]}`), func(t *testing.T, result string) {
		assert.Equal(t, "{}", result)
	})
}

func TestOctoPrintFileUpload(t *testing.T) {

	var err error
	files.DataDir, err = filepath.Abs("./testdata")
	if err != nil {
		t.Fatal(err)
	}

	files.Fs = afero.NewCopyOnWriteFs(afero.NewReadOnlyFs(afero.NewOsFs()), afero.NewMemMapFs())
	err = files.Init()
	if err != nil {
		t.Fatal(err)
	}

	marlinraker.State = marlinraker.Ready
	marlinraker.Config = config.DefaultConfig()
	notification.Testing = true

	testFileUpload(t, "/api/files/local", map[string]string{
		"root": "config",
		"path": "test/path",
	}, "file.txt", "testdata/test_upload.txt",
		func(t *testing.T, _ *httptest.ResponseRecorder, _ *any, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			filePath, err := filepath.Abs("testdata/config/test/path/file.txt")
			if err != nil {
				t.Fatal(err)
			}

			contents, err := afero.ReadFile(files.Fs, filePath)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, string(contents), "this is a test\n")
		})
}
