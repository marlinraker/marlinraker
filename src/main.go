package main

import (
	"context"
	_ "embed"
	"fmt"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/api"
	"marlinraker/src/config"
	"marlinraker/src/database"
	"marlinraker/src/files"
	"marlinraker/src/logger"
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/service"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func main() {
	os.Exit(run())
}

// run starts Marlinraker and returns the process exit code. Deferred cleanup
// runs before main calls os.Exit.
func run() int {
	exitCode := 0

	if os.Geteuid() == 0 {
		fmt.Println("Please do not run this program as root")
		return 1
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
	if err := os.WriteFile(pidFilePath, []byte(fmt.Sprint(os.Getpid())), 0644); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Remove(pidFilePath); err != nil {
			log.Errorf("Could not remove PID file: %v", err)
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
		log.Errorf("Unable to initialize database: %v", err)
		return 1
	}

	cfg, err := config.LoadConfig(filepath.Join(dataDir, "config/marlinraker.toml"))
	if err != nil {
		log.Errorf("Unable to load configuration: %v", err)
		return 1
	}

	if err := service.Init(cfg); err != nil {
		log.Warnf("Could not initialize service manager: %v", err)
	}
	defer service.Close()

	marlinraker.Init(cfg)

	server, serverErr := api.StartServer()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-ch:
		log.Printf("Received %s, shutting down", sig)
	case err := <-serverErr:
		log.Errorf("HTTP server failed: %v", err)
		exitCode = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Failed to shut down HTTP server: %v", err)
	}
	connections.TerminateAllConnections()
	return exitCode
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
		log.Errorf("Marlinraker is already running (PID: %s)", string(bytes))
		os.Exit(1)
	}
}
