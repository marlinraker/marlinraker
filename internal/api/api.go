package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/cors"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/api/executors"
	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/util"
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
	"machine.services.restart":      executors.MachineServicesRestart,
	"machine.services.start":        executors.MachineServicesStart,
	"machine.services.stop":         executors.MachineServicesStop,
	"machine.shutdown":              executors.MachineShutdown,
	"machine.system_info":           executors.MachineSystemInfo,
	"printer.emergency_stop":        executors.PrinterEmergencyStop,
	"printer.firmware_restart":      executors.PrinterFirmwareRestart,
	"printer.gcode.help":            executors.PrinterGcodeHelp,
	"printer.gcode.script":          executors.PrinterGcodeScript,
	"printer.info":                  executors.PrinterInfo,
	"printer.objects.list":          executors.PrinterObjectsList,
	"printer.objects.query":         executors.PrinterObjectsQuerySocket,
	"printer.objects.subscribe":     executors.PrinterObjectsSubscribeSocket,
	"printer.print.cancel":          executors.PrinterPrintCancel,
	"printer.print.pause":           executors.PrinterPrintPause,
	"printer.print.resume":          executors.PrinterPrintResume,
	"printer.print.start":           executors.PrinterPrintStart,
	"printer.restart":               executors.PrinterRestart,
	"server.config":                 executors.ServerConfig,
	"server.connection.identify":    executors.ServerConnectionIdentify,
	"server.database.delete_item":   executors.ServerDatabaseDeleteItem,
	"server.database.get_item":      executors.ServerDatabaseGetItem,
	"server.database.list":          executors.ServerDatabaseList,
	"server.database.post_item":     executors.ServerDatabasePostItem,
	"server.files.copy":             executors.ServerFilesCopy,
	"server.files.delete_directory": executors.ServerFilesDeleteDirectory,
	"server.files.delete_file":      executors.ServerFilesDeleteFile,
	"server.files.get_directory":    executors.ServerFilesGetDirectory,
	"server.files.list":             executors.ServerFilesList,
	"server.files.metadata":         executors.ServerFilesMetadata,
	"server.files.metascan":         executors.ServerFilesMetadata,
	"server.files.move":             executors.ServerFilesMove,
	"server.files.post_directory":   executors.ServerFilesPostDirectory,
	"server.files.roots":            executors.ServerFilesRoots,
	"server.files.thumbnails":       executors.ServerFilesThumbnails,
	"server.files.zip":              executors.ServerFilesZip,
	"server.gcode_store":            executors.ServerGcodeStore,
	"server.history.delete_job":     executors.ServerHistoryDeleteJob,
	"server.history.get_job":        executors.ServerHistoryGetJob,
	"server.history.list":           executors.ServerHistoryList,
	"server.history.reset_totals":   executors.ServerHistoryResetTotals,
	"server.history.totals":         executors.ServerHistoryTotals,
	"server.info":                   executors.ServerInfo,
	"server.logs.rollover":          executors.ServerLogsRollover,
	"server.restart":                executors.ServerRestart,
	"server.temperature_store":      executors.ServerTemperatureStore,
	"server.webcams.delete_item":    executors.ServerWebcamsDeleteItem,
	"server.webcams.list":           executors.ServerWebcamsList,
	"server.webcams.post_item":      executors.ServerWebcamsPostItem,
}

var httpExecutors = map[string]map[string]Executor{
	"GET": {
		"/access/oneshot_token":     executors.AccessOneshotToken,
		"/machine/proc_stats":       executors.MachineProcStats,
		"/machine/system_info":      executors.MachineSystemInfo,
		"/printer/gcode/help":       executors.PrinterGcodeHelp,
		"/printer/info":             executors.PrinterInfo,
		"/printer/objects/list":     executors.PrinterObjectsList,
		"/printer/objects/query":    executors.PrinterObjectsQueryHttp,
		"/server/config":            executors.ServerConfig,
		"/server/database/item":     executors.ServerDatabaseGetItem,
		"/server/database/list":     executors.ServerDatabaseList,
		"/server/files/directory":   executors.ServerFilesGetDirectory,
		"/server/files/list":        executors.ServerFilesList,
		"/server/files/metadata":    executors.ServerFilesMetadata,
		"/server/files/metascan":    executors.ServerFilesMetascan,
		"/server/files/roots":       executors.ServerFilesRoots,
		"/server/files/thumbnails":  executors.ServerFilesThumbnails,
		"/server/gcode_store":       executors.ServerGcodeStore,
		"/server/history/job":       executors.ServerHistoryGetJob,
		"/server/history/list":      executors.ServerHistoryList,
		"/server/history/totals":    executors.ServerHistoryTotals,
		"/server/info":              executors.ServerInfo,
		"/server/temperature_store": executors.ServerTemperatureStore,
		"/server/webcams/item":      executors.ServerWebcamsGetItem,
		"/server/webcams/list":      executors.ServerWebcamsList,
	},
	"POST": {
		"/machine/reboot":              executors.MachineReboot,
		"/machine/services/restart":    executors.MachineServicesRestart,
		"/machine/services/start":      executors.MachineServicesStart,
		"/machine/services/stop":       executors.MachineServicesStop,
		"/machine/shutdown":            executors.MachineShutdown,
		"/printer/emergency_stop":      executors.PrinterEmergencyStop,
		"/printer/firmware_restart":    executors.PrinterFirmwareRestart,
		"/printer/gcode/script":        executors.PrinterGcodeScript,
		"/printer/objects/subscribe":   executors.PrinterObjectsSubscribeHttp,
		"/printer/print/cancel":        executors.PrinterPrintCancel,
		"/printer/print/pause":         executors.PrinterPrintPause,
		"/printer/print/resume":        executors.PrinterPrintResume,
		"/printer/print/start":         executors.PrinterPrintStart,
		"/printer/restart":             executors.PrinterRestart,
		"/server/database/item":        executors.ServerDatabasePostItem,
		"/server/files/copy":           executors.ServerFilesCopy,
		"/server/files/directory":      executors.ServerFilesPostDirectory,
		"/server/files/move":           executors.ServerFilesMove,
		"/server/files/upload":         executors.ServerFilesUpload,
		"/server/files/zip":            executors.ServerFilesZip,
		"/server/history/reset_totals": executors.ServerHistoryResetTotals,
		"/server/logs/rollover":        executors.ServerLogsRollover,
		"/server/restart":              executors.ServerRestart,
		"/server/webcams/item":         executors.ServerWebcamsPostItem,
	},
	"DELETE": {
		"/server/database/item":   executors.ServerDatabaseDeleteItem,
		"/server/files/directory": executors.ServerFilesDeleteDirectory,
		"/server/history/job":     executors.ServerHistoryDeleteJob,
		"/server/webcams/item":    executors.ServerWebcamsDeleteItem,
	},
}

type HttpHandler struct{}

func (HttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	log.WithField("method", request.Method).Debugln(request.URL.String())

	requestPath := strings.TrimRight(request.URL.Path, "/")
	err := handlePath(writer, request, requestPath)
	if err != nil {
		log.Errorf("Error while handling %s %s: %v", request.Method, requestPath, err)
		writer.WriteHeader(500)
		_, _ = writer.Write([]byte(err.Error()))
	}
}

func handlePath(writer http.ResponseWriter, request *http.Request, requestPath string) error {
	switch {

	case requestPath == "":
		return handleIndex(writer, request)

	case requestPath == "/websocket":
		return handleSocket(writer, request)

	case marlinraker.Config.Misc.OctoprintCompat && strings.HasPrefix(requestPath, "/api/"):
		return handleOctoPrint(writer, request)

	case isFilePath(requestPath) && (request.Method == "GET" || request.Method == "DELETE"):
		if request.Method == "GET" {
			return handleFileDownload(writer, request)
		} else {
			return handleFileDelete(writer, request)
		}

	default:
		return handleHttp(writer, request)
	}
}

// origins decides which browser origins may use the API. It is replaced by
// StartServer with the configured cors_domains.
var origins = newOriginChecker(nil)

// StartServer starts the HTTP server in the background. The returned channel
// receives the error if the server stops unexpectedly.
func StartServer() (*http.Server, <-chan error) {
	address := fmt.Sprintf("%s:%d", marlinraker.Config.Web.BindAddress, marlinraker.Config.Web.Port)
	origins = newOriginChecker(marlinraker.Config.Web.CorsDomains)

	corsHandler := cors.New(cors.Options{
		AllowOriginVaryRequestFunc: func(request *http.Request, origin string) (bool, []string) {
			return origins.isAllowed(request, origin), nil
		},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead},
		AllowedHeaders: []string{"*"},
	})

	server := &http.Server{
		Addr:              address,
		Handler:           origins.handler(corsHandler.Handler(HttpHandler{})),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Listening on %s", address)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()
	return server, errCh
}

func isFilePath(path string) bool {
	if lo.Contains([]string{"/server/files/klippy.log", "/server/files/moonraker.log", "/server/files/marlinraker.log"}, path) {
		return true
	}
	_, ok := fileRootOf(path)
	return ok
}

// fileRootOf returns the file root a /server/files/... request path points
// into, after resolving "..". It returns false if the path does not stay inside
// a registered file root.
func fileRootOf(requestPath string) (files.FileRoot, bool) {
	relPath, found := strings.CutPrefix(requestPath, "/server/files/")
	if !found {
		return files.FileRoot{}, false
	}
	rootName, _, _ := strings.Cut(util.SanitizePath(relPath), "/")
	return lo.Find(files.FileRoots, func(root files.FileRoot) bool {
		return root.Name == rootName
	})
}
