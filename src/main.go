package main

import (
	_ "embed"
	"fmt"
	log "github.com/sirupsen/logrus"
	"marlinraker-go/src/api"
	"marlinraker-go/src/files"
	"marlinraker-go/src/logger"
	"marlinraker-go/src/marlinraker"
	"os"
	"os/signal"
	"path/filepath"
)

func main() {

	if os.Geteuid() == 0 {
		fmt.Println("Please do not run this program as root")
		return
	}

	dataDir, err := filepath.Abs("./marlinraker_files")
	if err != nil {
		panic(err)
	}

	files.DataDir = dataDir
	err = files.CreateFileRoots()
	if err != nil {
		panic(err)
	}

	logFile, err := logger.SetupLogger(dataDir)
	if err != nil {
		panic(err)
	}
	defer logger.CloseLogger(logFile)

	marlinraker.Init(dataDir)

	go api.StartServer()

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	<-ch
	log.Println("Received interrupt, shutting down")
}
