package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"io"
	"marlinraker/src/files"
	"os"
	"path/filepath"
)

func SetupLogger(dataDir string) (afero.File, error) {

	logFilePath := filepath.Join(dataDir, "logs/marlinraker.log")
	logFile, err := files.Fs.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}

	log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        true,
	})

	return logFile, nil
}

func CloseLogger(logFile afero.File) {
	if err := logFile.Close(); err != nil {
		panic(err)
	}
}
