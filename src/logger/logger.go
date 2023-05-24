package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"io"
	"marlinraker/src/files"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

var (
	logFile afero.File
	mu      = &sync.Mutex{}
)

func SetupLogger() error {

	mu.Lock()
	defer mu.Unlock()

	if err := openLogFile(); err != nil {
		return err
	}

	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        true,
	})

	return nil
}

func HandleRotate() {

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGUSR1)

	for {
		<-ch
		log.Println("Log rotate requested")
		mu.Lock()
		if err := openLogFile(); err != nil {
			panic(err)
		}
		mu.Unlock()
	}
}

func CloseLogger() error {
	mu.Lock()
	defer mu.Unlock()
	return logFile.Close()
}

func openLogFile() error {

	var err error
	if logFile != nil {
		if err = logFile.Close(); err != nil {
			return err
		}
	}

	logFilePath := filepath.Join(files.DataDir, "logs/marlinraker.log")
	logFile, err = files.Fs.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	return nil
}
