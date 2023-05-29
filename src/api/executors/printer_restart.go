package executors

import (
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
	"syscall"
)

func PrinterRestart(*connections.Connection, *http.Request, Params) (any, error) {
	go func() {
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
			util.LogError(err)
		}
	}()
	return "ok", nil
}
