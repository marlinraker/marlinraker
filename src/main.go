package main

import (
	_ "embed"
	"fmt"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/api"
	"marlinraker/src/database"
	"marlinraker/src/files"
	"marlinraker/src/logger"
	"marlinraker/src/marlinraker"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {

	if os.Geteuid() == 0 {
		fmt.Println("Please do not run this program as root")
		return
	}

	relPath := os.Getenv("MARLINRAKER_DIR")
	if relPath == "" {
		relPath = "./marlinraker_files"
	}
	dataDir, err := filepath.Abs(relPath)
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(err)
	}

	pidFilePath := filepath.Join(dataDir, "marlinraker.pid")
	checkAlreadyRunning(pidFilePath)
	if err := os.WriteFile(pidFilePath, []byte(strconv.Itoa(os.Getpid())), 0755); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Remove(pidFilePath); err != nil {
			panic(err)
		}
	}()

	files.DataDir = dataDir
	if err := files.Init(); err != nil {
		panic(err)
	}

	if err := logger.SetupLogger(); err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.CloseLogger(); err != nil {
			panic(err)
		}
	}()
	go logger.HandleRotate()

	if err := database.Init(); err != nil {
		panic(err)
	}

	marlinraker.Init(dataDir)

	go api.StartServer()

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	<-ch
	log.Println("Received interrupt, shutting down")
}

func checkAlreadyRunning(pidFilePath string) {
	bytes, err := os.ReadFile(pidFilePath)
	if err != nil {
		return
	}

	pid, err := strconv.Atoi(string(bytes))
	if err != nil {
		return
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	if process.Signal(syscall.Signal(0)) == nil {
		panic("marlinraker is already running (" + string(bytes) + ")")
	}
}
