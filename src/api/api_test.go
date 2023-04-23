package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gorilla/websocket"
	"github.com/spf13/afero"
	"gotest.tools/assert"
	"marlinraker-go/src/api/executors"
	"marlinraker-go/src/api/notification"
	"marlinraker-go/src/config"
	"marlinraker-go/src/constants"
	"marlinraker-go/src/database"
	"marlinraker-go/src/files"
	"marlinraker-go/src/marlinraker"
	"marlinraker-go/src/marlinraker/gcode_store"
	"marlinraker-go/src/marlinraker/temp_store"
	"marlinraker-go/src/printer_objects"
	"marlinraker-go/src/system_info"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func testHttp[Result any](t *testing.T, method string, endpoint string, params executors.Params,
	f func(*testing.T, *httptest.ResponseRecorder, *Result, *Error)) {
	t.Run(method+endpoint, func(t *testing.T) {

		urlQuery, err := url.Parse(endpoint)
		if err != nil {
			t.Fatal(err)
		}

		values := urlQuery.Query()
		for param, value := range params {
			values.Set(param, fmt.Sprintf("%v", value))
		}
		urlQuery.RawQuery = values.Encode()

		request, _ := http.NewRequest(method, urlQuery.String(), nil)
		recorder := httptest.NewRecorder()
		handleHttp(recorder, request)

		var errorResponse ErrorResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatal(err)
			return
		}
		if errorResponse.Error.Code > 0 {
			f(t, recorder, nil, &errorResponse.Error)
			return
		}

		var response struct {
			Result Result `json:"result"`
		}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
			return
		}
		f(t, recorder, &response.Result, nil)
	})
}

func testSocket[Result any](t *testing.T, method string, params executors.Params, f func(*testing.T, *Result, *Error)) {
	t.Run(method, func(t *testing.T) {

		server := httptest.NewServer(http.HandlerFunc(handleSocket))
		defer server.Close()
		socketUrl := "ws" + server.URL[4:]

		socket, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func(socket *websocket.Conn) {
			_ = socket.Close()
		}(socket)

		request := RpcRequest{
			Rpc: Rpc{
				JsonRpc: "2.0",
				Id:      0,
			},
			Method: method,
			Params: params,
		}

		{
			data, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}

			err = socket.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			_, data, err := socket.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}

			var errorResponse RpcErrorResponse
			err = json.Unmarshal(data, &errorResponse)
			if err != nil {
				t.Fatal(err)
			}
			if errorResponse.Error.Code > 0 {
				f(t, nil, &errorResponse.Error)
				return
			}

			var response struct {
				Result Result `json:"result"`
			}
			err = json.Unmarshal(data, &response)
			if err != nil {
				t.Fatal(err)
			}
			f(t, &response.Result, nil)
		}
	})
}

func testAll[Result any](t *testing.T, rpcMethod string, httpMethod string, httpUrl string, params executors.Params,
	f func(*testing.T, *httptest.ResponseRecorder, *Result, *Error)) {

	testSocket(t, rpcMethod, params, func(t *testing.T, result *Result, error *Error) {
		f(t, nil, result, error)
	})
	testHttp(t, httpMethod, httpUrl, params, f)
}

func makeConnection(t *testing.T) (*websocket.Conn, int) {

	server := httptest.NewServer(http.HandlerFunc(handleSocket))
	defer server.Close()
	socketUrl := "ws" + server.URL[4:]

	socket, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = socket.WriteJSON(RpcRequest{
		Rpc:    Rpc{JsonRpc: "2.0", Id: 42},
		Method: "server.connection.identify",
		Params: executors.Params{"client_name": "", "version": "", "type": "", "url": ""},
	})
	if err != nil {
		t.Fatal(err)
	}

	var response RpcResultResponse
	err = socket.ReadJSON(&response)
	if err != nil {
		t.Fatal(err)
	}

	connectionId := response.Result.(map[string]any)["connection_id"].(float64)
	return socket, int(connectionId)
}

type TestObject struct{}

func (TestObject) Query() printer_objects.QueryResult {
	return map[string]any{
		"attribute1": "value1",
		"attribute2": "value2",
	}
}

func TestApi(t *testing.T) {

	var err error
	files.DataDir, err = filepath.Abs("./testdata")
	if err != nil {
		t.Fatal(err)
	}

	files.Fs = afero.NewCopyOnWriteFs(afero.NewReadOnlyFs(afero.NewOsFs()), afero.NewMemMapFs())
	err = files.CreateFileRoots()
	if err != nil {
		t.Fatal(err)
	}

	marlinraker.State = marlinraker.Ready
	marlinraker.Config = config.DefaultConfig()
	notification.Testing = true

	err = database.Init()
	if err != nil {
		t.Fatal(err)
	}

	gcode_store.LogNow("test command", gcode_store.Command)
	gcode_store.LogNow("test response", gcode_store.Response)

	systemInfo, err := system_info.GetSystemInfo()
	if err != nil {
		t.Fatal(err)
	}

	printer_objects.RegisterObject("test_object", TestObject{})

	socket, cid := makeConnection(t)
	defer func(socket *websocket.Conn) {
		_ = socket.Close()
	}(socket)

	testAll(t, "server.info", "GET", "/server/info", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerInfoResult, error *Error) {

			if response != nil {
				assert.Equal(t, response.Code, 200)
			}

			if error != nil {
				t.Fatal(error)
			}

			assert.DeepEqual(t, result, &executors.ServerInfoResult{
				KlippyConnected:           true,
				KlippyState:               "ready",
				Components:                []string{"server", "file_manager", "machine", "database", "data_store", "proc_stats", "history"},
				FailedComponents:          []string{},
				RegisteredDirectories:     files.GetRegisteredDirectories(),
				Warnings:                  []string{},
				WebsocketCount:            0,
				MissingKlippyRequirements: []string{},
				MoonrakerVersion:          constants.Version,
				ApiVersion:                constants.ApiVersion,
				ApiVersionString:          constants.ApiVersionString,
				Type:                      "marlinraker",
			}, cmpopts.IgnoreFields(executors.ServerInfoResult{}, "WebsocketCount"))
		})

	testAll(t, "printer.objects.list", "GET", "/printer/objects/list", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.PrinterObjectsListResult, error *Error) {

			if response != nil {
				assert.Equal(t, response.Code, 200)
			}

			if error != nil {
				t.Fatal(error)
			}

			assert.DeepEqual(t, result, &executors.PrinterObjectsListResult{
				Objects: []string{"test_object"},
			})
		})

	testSocket(t, "server.connection.identify", executors.Params{
		"client_name": "test",
		"version":     "0.1.0",
		"type":        "web",
		"url":         "example.com",
	}, func(t *testing.T, result *executors.ServerConnectionIdentifyResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &executors.ServerConnectionIdentifyResult{},
			cmpopts.IgnoreFields(executors.ServerConnectionIdentifyResult{}, "ConnectionId"))
	})

	testSocket(t, "server.connection.identify", executors.Params{
		"client_name": "test",
		"type":        "web",
		"url":         "example.com",
	}, func(t *testing.T, result *executors.ServerConnectionIdentifyResult, error *Error) {

		assert.Check(t, error != nil)
		assert.DeepEqual(t, error, &Error{Code: 400, Message: "version param is required"})
	})

	testAll(t, "server.files.list", "GET", "/server/files/list", executors.Params{
		"root": "config",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerFilesResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &[]files.File{
			{Path: "foo/bar.txt", Permissions: "rw"},
			{Path: "foobar.txt", Permissions: "rw"},
		}, cmpopts.IgnoreFields(files.File{}, "Modified"))
	})

	testAll(t, "server.config", "GET", "/server/config", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerConfigResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			assert.DeepEqual(t, result.Config, marlinraker.Config)
		})

	testAll(t, "machine.system_info", "GET", "/machine/system_info", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.MachineSystemInfoResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			assert.DeepEqual(t, result.SystemInfo, systemInfo)
		})

	testAll(t, "machine.proc_stats", "GET", "/machine/proc_stats", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.MachineProcStatsResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			assert.DeepEqual(t, result, system_info.GetStats())
		})

	testAll(t, "server.database.list", "GET", "/server/database/list", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerDatabaseListResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			assert.DeepEqual(t, result, &executors.ServerDatabaseListResult{
				Namespaces: []string{"namespace_1", "namespace_2"},
			}, cmpopts.SortSlices(func(a string, b string) bool { return a < b }))
		})

	testAll(t, "printer.info", "GET", "/printer/info", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.PrinterInfoResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			hostname, err := os.Hostname()
			if err != nil {
				t.Fatal(err)
			}

			systemInfo, err := system_info.GetSystemInfo()
			if err != nil {
				t.Fatal(err)
			}

			assert.DeepEqual(t, result, &executors.PrinterInfoResult{
				State:           marlinraker.State,
				StateMessage:    marlinraker.StateMessage,
				Hostname:        hostname,
				SoftwareVersion: constants.Version,
				CpuInfo:         systemInfo.CpuInfo.CpuDesc,
			})
		})

	testAll(t, "server.gcode_store", "GET", "/server/gcode_store", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerGcodeStoreResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			assert.DeepEqual(t, result.GcodeStore, gcode_store.GcodeStore)
		})

	testAll(t, "server.temperature_store", "GET", "/server/temperature_store", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerTemperatureStoreResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			store := temp_store.GetStore()
			assert.DeepEqual(t, (*temp_store.TempStore)(result), &store)
		})

	testAll(t, "printer.objects.query", "GET", "/printer/objects/query", executors.Params{
		"test_object": "",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.PrinterObjectsQueryResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result.Status, map[string]printer_objects.QueryResult{
			"test_object": {
				"attribute1": "value1",
				"attribute2": "value2",
			},
		})
	})

	testAll(t, "printer.objects.query", "GET", "/printer/objects/query", executors.Params{
		"test_object": "attribute1,attribute3",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.PrinterObjectsQueryResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result.Status, map[string]printer_objects.QueryResult{
			"test_object": {
				"attribute1": "value1",
			},
		})
	})

	testSocket(t, "printer.objects.subscribe", executors.Params{
		"objects": map[string]any{"test_object": nil},
	}, func(t *testing.T, result *executors.PrinterObjectsSubscribeResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result.Status, map[string]printer_objects.QueryResult{
			"test_object": {
				"attribute1": "value1",
				"attribute2": "value2",
			},
		})
	})

	testSocket(t, "printer.objects.subscribe", executors.Params{
		"objects": map[string]any{"test_object": []string{"attribute1"}},
	}, func(t *testing.T, result *executors.PrinterObjectsSubscribeResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result.Status, map[string]printer_objects.QueryResult{
			"test_object": {
				"attribute1": "value1",
			},
		})
	})

	testHttp(t, "POST", "/printer/objects/subscribe", executors.Params{
		"connection_id": cid,
		"test_object":   "",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.PrinterObjectsSubscribeResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result.Status, map[string]printer_objects.QueryResult{
			"test_object": {
				"attribute1": "value1",
				"attribute2": "value2",
			},
		})
	})

	testHttp(t, "POST", "/printer/objects/subscribe", executors.Params{
		"connection_id": cid,
		"test_object":   "attribute1",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.PrinterObjectsSubscribeResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result.Status, map[string]printer_objects.QueryResult{
			"test_object": {
				"attribute1": "value1",
			},
		})
	})

	testAll(t, "printer.gcode.script", "POST", "/printer/gcode/script", executors.Params{
		"script": "G28",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *string, error *Error) {

		if result != nil {
			fmt.Printf("%v\n", result)
			t.Fatal(result)
		}

		assert.Equal(t, error.Message, "printer not online")
	})

	testAll(t, "server.files.roots", "GET", "/server/files/roots", executors.Params{},
		func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerFilesRootsResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			assert.DeepEqual(t, result, (*executors.ServerFilesRootsResult)(&files.FileRoots))
		})

	testAll(t, "server.files.get_directory", "GET", "/server/files/directory", executors.Params{
		"path": "config",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerFilesGetDirectoryResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &executors.ServerFilesGetDirectoryResult{
			Dirs:     []files.DirectoryMeta{{DirName: "foo", Permissions: "rw"}},
			Files:    []files.FileMeta{{FileName: "foobar.txt", Permissions: "rw"}},
			RootInfo: files.RootInfo{Name: "config", Permissions: "rw"},
		},
			cmpopts.IgnoreFields(executors.ServerFilesGetDirectoryResult{}, "DiskUsage"),
			cmpopts.IgnoreFields(files.DirectoryMeta{}, "Modified", "Size"),
			cmpopts.IgnoreFields(files.FileMeta{}, "Modified", "Size"),
		)
	})

	testAll(t, "server.files.post_directory", "POST", "/server/files/directory", executors.Params{
		"path": "config/bar",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerFilesPostDirectoryResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &executors.ServerFilesPostDirectoryResult{
			Item: files.ActionItem{
				Path:        "bar",
				Root:        "config",
				Permissions: "rw",
			},
			Action: "create_dir",
		}, cmpopts.IgnoreFields(files.ActionItem{}, "Modified", "Size"))
	})

	testAll(t, "server.files.delete_directory", "DELETE", "/server/files/directory", executors.Params{
		"path": "config/foo",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerFilesPostDirectoryResult, error *Error) {

		assert.Equal(t, error.Message, "directory is not empty")
	})

	testAll(t, "server.files.delete_directory", "DELETE", "/server/files/directory", executors.Params{
		"path":  "config/foo",
		"force": "true",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerFilesPostDirectoryResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &executors.ServerFilesPostDirectoryResult{
			Item: files.ActionItem{
				Path:        "foo",
				Root:        "config",
				Permissions: "rw",
			},
			Action: "delete_dir",
		}, cmpopts.IgnoreFields(files.ActionItem{}, "Modified", "Size"))
	})

	// create virtual file for move
	moveSourcePath := filepath.Join(files.DataDir, "config/foo/test.txt")
	_, err = files.Fs.Create(moveSourcePath)
	if err != nil {
		t.Fatal(err)
	}

	testAll(t, "server.files.move", "POST", "/server/files/move", executors.Params{
		"source": "config/foo/test.txt",
		"dest":   "config/test.txt",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerFilesMoveResult, error *Error) {

		// re-create file
		_, err = files.Fs.Create(moveSourcePath)
		if err != nil {
			t.Fatal(err)
		}

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &executors.ServerFilesMoveResult{
			Item: files.ActionItem{
				Path:        "test.txt",
				Root:        "config",
				Permissions: "rw",
			},
			SourceItem: files.ActionItem{
				Path: "foo/test.txt",
				Root: "config",
			},
			Action: "move_file",
		}, cmpopts.IgnoreFields(files.ActionItem{}, "Modified", "Size"))
	})
}
