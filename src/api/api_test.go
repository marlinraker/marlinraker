package api

import (
	"encoding/base32"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gorilla/websocket"
	"github.com/spf13/afero"
	"golang.org/x/exp/slices"
	"gotest.tools/assert"
	"marlinraker/src/api/executors"
	"marlinraker/src/api/notification"
	"marlinraker/src/config"
	"marlinraker/src/constants"
	"marlinraker/src/database"
	"marlinraker/src/files"
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/gcode_store"
	"marlinraker/src/marlinraker/temp_store"
	"marlinraker/src/printer_objects"
	"marlinraker/src/system_info"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

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
				Namespaces: append(database.ReservedNamespaces, "namespace_1", "namespace_2"),
			}, cmpopts.SortSlices(func(a string, b string) bool { return a < b }))
		})

	testAll(t, "server.database.get_item", "GET", "/server/database/item", executors.Params{
		"namespace": "namespace_1",
		"key":       "key1",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerDatabaseGetItemResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		key := "key1"
		assert.DeepEqual(t, result, &executors.ServerDatabaseGetItemResult{
			Namespace: "namespace_1",
			Key:       &key,
			Value:     "string",
		})
	})

	testAll(t, "server.database.get_item", "GET", "/server/database/item", executors.Params{
		"namespace": "namespace_1",
		"key":       "key2",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerDatabaseGetItemResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		key := "key2"
		assert.DeepEqual(t, result, &executors.ServerDatabaseGetItemResult{
			Namespace: "namespace_1",
			Key:       &key,
			Value:     123.0,
		})
	})

	testAll(t, "server.database.get_item", "GET", "/server/database/item", executors.Params{
		"namespace": "namespace_1",
		"key":       "key3",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerDatabaseGetItemResult, error *Error) {
		assert.Error(t, error, "Key \"key3\" in namespace \"namespace_1\" not found")
	})

	testAll(t, "server.database.post_item", "POST", "/server/database/item", executors.Params{
		"namespace": "namespace_1",
		"key":       "key3",
		"value":     true,
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerDatabasePostItemResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &executors.ServerDatabasePostItemResult{
			Namespace: "namespace_1",
			Key:       "key3",
			Value:     true,
		})
	})

	// create entry to delete
	if _, err := database.PostItem("test_namespace", "delete_me", "delete_me"); err != nil {
		t.Fatal(err)
	}

	testAll(t, "server.database.delete_item", "DELETE", "/server/database/item", executors.Params{
		"namespace": "test_namespace",
		"key":       "delete_me",
	}, func(t *testing.T, response *httptest.ResponseRecorder, result *executors.ServerDatabaseDeleteItemResult, error *Error) {

		_, err := database.GetItem("test_namespace", "delete_me")
		assert.Error(t, err, "Key \"delete_me\" in namespace \"test_namespace\" not found")

		namespaces := database.ListNamespaces()
		assert.Equal(t, slices.Contains(namespaces, "delete_me"), false)

		// re-create entry
		if _, err := database.PostItem("test_namespace", "delete_me", "delete_me"); err != nil {
			t.Fatal(err)
		}

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &executors.ServerDatabaseDeleteItemResult{
			Namespace: "test_namespace",
			Key:       "delete_me",
			Value:     "delete_me",
		})
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

	testSocket(t, "printer.objects.query", executors.Params{
		"objects": map[string]any{
			"test_object": nil,
		},
	}, func(t *testing.T, result *executors.PrinterObjectsQueryResult, error *Error) {

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

	testHttp(t, "GET", "/printer/objects/query", executors.Params{
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

	testSocket(t, "printer.objects.query", executors.Params{
		"objects": map[string]any{
			"test_object": []string{"attribute1", "attribute3"},
		},
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

	testHttp(t, "GET", "/printer/objects/query", executors.Params{
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
		"objects": map[string]any{
			"test_object": nil,
		},
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

	testSocket(t, "printer.objects.subscribe", executors.Params{
		"objects": map[string]any{
			"test_object": []string{"attribute1"},
		},
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
			t.Fatal(result)
		}

		assert.Equal(t, error.Message, "printer is not online")
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

	testFileUpload(t, "/server/files/upload", map[string]string{
		"root": "config",
		"path": "test/path",
	}, "file.txt", "testdata/test_upload.txt",
		func(t *testing.T, _ *httptest.ResponseRecorder, result *executors.ServerFilesUploadResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			assert.DeepEqual(t, result, &executors.ServerFilesUploadResult{
				Item: files.ActionItem{
					Path:        "test/path/file.txt",
					Root:        "config",
					Size:        15,
					Permissions: "rw",
				},
				Action: "create_file",
			}, cmpopts.IgnoreFields(files.ActionItem{}, "Modified"))

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

	// create test file for deletion
	testFilePath := filepath.Join(files.DataDir, "config/delete_me")
	if err := afero.WriteFile(files.Fs, testFilePath, []byte("delete me"), 0755); err != nil {
		t.Fatal(err)
	}

	testAll(t, "server.files.delete_file", "DELETE", "/server/files/config/delete_me", executors.Params{
		"path": "config/delete_me",
	}, func(t *testing.T, _ *httptest.ResponseRecorder, result *executors.ServerFilesDeleteFileResult, error *Error) {

		// recreate file
		if err := afero.WriteFile(files.Fs, testFilePath, []byte("delete me"), 0755); err != nil {
			t.Fatal(err)
		}

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &executors.ServerFilesDeleteFileResult{
			Item: files.ActionItem{
				Path:        "delete_me",
				Root:        "config",
				Permissions: "",
				Size:        0,
				Modified:    0,
			},
			Action: "delete_file",
		})
	})

	testAll(t, "server.files.zip", "POST", "/server/files/zip", executors.Params{
		"dest": "config/archive.zip",
		"items": []string{
			"config/foo",
			"config/foobar.txt",
		},
		"store_only": false,
	}, func(t *testing.T, _ *httptest.ResponseRecorder, result *executors.ServerFilesZipResult, error *Error) {

		if error != nil {
			t.Fatal(error)
		}

		assert.DeepEqual(t, result, &executors.ServerFilesZipResult{
			Destination: files.ActionItem{
				Path:        "archive.zip",
				Root:        "config",
				Permissions: "rw",
			},
			Action: "zip_files",
		}, cmpopts.IgnoreFields(files.ActionItem{}, "Modified", "Size"))
	})

	testAll(t, "access.oneshot_token", "GET", "/access/oneshot_token", executors.Params{},
		func(t *testing.T, _ *httptest.ResponseRecorder, result *executors.AccessOneshotTokenResult, error *Error) {

			if error != nil {
				t.Fatal(error)
			}

			buf, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(string(*result))
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, len(buf), 20)
		})
}
