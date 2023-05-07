package api

import (
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/api/executors"
	"marlinraker/src/files"
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"net/http"
	"strconv"
	"strings"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err Error) Error() string {
	return err.Message
}

type Executor func(*connections.Connection, *http.Request, executors.Params) (any, error)

var socketExecutors = map[string]Executor{
	"access.oneshot_token":          executors.AccessOneshotToken,
	"machine.proc_stats":            executors.MachineProcStats,
	"machine.reboot":                executors.MachineReboot,
	"machine.shutdown":              executors.MachineShutdown,
	"machine.system_info":           executors.MachineSystemInfo,
	"printer.firmware_restart":      executors.PrinterFirmwareRestart,
	"printer.gcode.help":            executors.PrinterGcodeHelp,
	"printer.gcode.script":          executors.PrinterGcodeScript,
	"printer.info":                  executors.PrinterInfo,
	"printer.objects.list":          executors.PrinterObjectsList,
	"printer.objects.query":         executors.PrinterObjectsQuery,
	"printer.objects.subscribe":     executors.PrinterObjectsSubscribeSocket,
	"server.config":                 executors.ServerConfig,
	"server.connection.identify":    executors.ServerConnectionIdentify,
	"server.database.delete_item":   executors.ServerDatabaseDeleteItem,
	"server.database.get_item":      executors.ServerDatabaseGetItem,
	"server.database.list":          executors.ServerDatabaseList,
	"server.database.post_item":     executors.ServerDatabasePostItem,
	"server.files.delete_directory": executors.ServerFilesDeleteDirectory,
	"server.files.delete_file":      executors.ServerFilesDeleteFile,
	"server.files.get_directory":    executors.ServerFilesGetDirectory,
	"server.files.list":             executors.ServerFilesList,
	"server.files.move":             executors.ServerFilesMove,
	"server.files.post_directory":   executors.ServerFilesPostDirectory,
	"server.files.roots":            executors.ServerFilesRoots,
	"server.files.zip":              executors.ServerFilesZip,
	"server.gcode_store":            executors.ServerGcodeStore,
	"server.history.list":           executors.ServerHistoryList,
	"server.info":                   executors.ServerInfo,
	"server.temperature_store":      executors.ServerTemperatureStore,
}

var httpExecutors = map[string]map[string]Executor{
	"GET": {
		"/access/oneshot_token":     executors.AccessOneshotToken,
		"/machine/proc_stats":       executors.MachineProcStats,
		"/machine/system_info":      executors.MachineSystemInfo,
		"/printer/gcode/help":       executors.PrinterGcodeHelp,
		"/printer/info":             executors.PrinterInfo,
		"/printer/objects/list":     executors.PrinterObjectsList,
		"/printer/objects/query":    executors.PrinterObjectsQuery,
		"/server/config":            executors.ServerConfig,
		"/server/database/item":     executors.ServerDatabaseGetItem,
		"/server/database/list":     executors.ServerDatabaseList,
		"/server/files/directory":   executors.ServerFilesGetDirectory,
		"/server/files/list":        executors.ServerFilesList,
		"/server/files/roots":       executors.ServerFilesRoots,
		"/server/gcode_store":       executors.ServerGcodeStore,
		"/server/history/list":      executors.ServerHistoryList,
		"/server/info":              executors.ServerInfo,
		"/server/temperature_store": executors.ServerTemperatureStore,
	},
	"POST": {
		"/machine/reboot":            executors.MachineReboot,
		"/machine/shutdown":          executors.MachineShutdown,
		"/printer/firmware_restart":  executors.PrinterFirmwareRestart,
		"/printer/gcode/script":      executors.PrinterGcodeScript,
		"/printer/objects/subscribe": executors.PrinterObjectsSubscribeHttp,
		"/server/database/item":      executors.ServerDatabasePostItem,
		"/server/files/directory":    executors.ServerFilesPostDirectory,
		"/server/files/move":         executors.ServerFilesMove,
		"/server/files/upload":       executors.ServerFilesUpload,
		"/server/files/zip":          executors.ServerFilesZip,
	},
	"DELETE": {
		"/server/database/item":   executors.ServerDatabaseDeleteItem,
		"/server/files/directory": executors.ServerFilesDeleteDirectory,
	},
}

type HttpHandler struct{}

func (HttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	log.WithField("method", request.Method).Debugln(request.URL.String())

	requestPath := strings.TrimRight(request.URL.Path, "/")
	switch {

	case requestPath == "":
		if err := handleIndex(writer, request); err != nil {
			log.Error(err)
		}

	case requestPath == "/websocket":
		if err := handleSocket(writer, request); err != nil {
			log.Error(err)
		}

	case isFilePath(requestPath) && (request.Method == "GET" || request.Method == "DELETE"):
		if request.Method == "GET" {
			handleFileDownload(writer, request)
		} else if err := handleFileDelete(writer, request); err != nil {
			log.Error(err)
		}

	default:
		if err := handleHttp(writer, request); err != nil {
			log.Error(err)
		}
	}
}

func StartServer() {
	address := marlinraker.Config.Web.BindAddress + ":" + strconv.Itoa(marlinraker.Config.Web.Port)
	log.Println("Listening on " + address)

	err := http.ListenAndServe(address, cors.AllowAll().Handler(HttpHandler{}))
	if err != nil {
		panic(err)
	}
}

func isFilePath(path string) bool {
	for _, root := range files.FileRoots {
		if strings.HasPrefix(path, "/server/files/"+root.Name) {
			return true
		}
	}
	return false
}
