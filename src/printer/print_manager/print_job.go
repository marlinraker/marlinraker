package print_manager

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/files"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type printJob struct {
	manager        *PrintManager
	fileName       string
	filePath       string
	pauseMutex     *sync.Mutex
	isPaused       bool
	isStarted      bool
	hasEnded       bool
	isCanceled     bool
	scanner        *bufio.Scanner
	position       int64
	fileSize       int64
	progress       float32
	startTime      time.Time
	lastResumeTime time.Time
	endTime        time.Time
	printDuration  time.Duration
}

func newPrintJob(manager *PrintManager, fileName string) *printJob {
	filePath := filepath.Join(files.DataDir, "gcodes", fileName)
	job := &printJob{
		manager:    manager,
		fileName:   fileName,
		filePath:   filePath,
		pauseMutex: &sync.Mutex{},
		isPaused:   false,
		isStarted:  false,
		hasEnded:   false,
		isCanceled: false,
		position:   0,
		fileSize:   0,
		progress:   0,
	}
	return job
}

func (job *printJob) start() error {

	stat, err := files.Fs.Stat(job.filePath)
	if err != nil {
		return err
	}

	reader, err := files.Fs.OpenFile(job.filePath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	now := time.Now()
	job.startTime = now
	job.lastResumeTime = now
	job.isStarted = true
	job.scanner = bufio.NewScanner(reader)
	job.fileSize = stat.Size()
	job.printDuration = 0
	job.position = 0
	job.progress = 0

	go func() {
		defer func() {
			if err := reader.Close(); err != nil {
				log.Error(err)
			}
		}()

		for job.scanner.Scan() {
			if job.isCanceled {
				return
			}
			if read, err := job.nextLine(); err != nil {
				log.Error(err)
				return
			} else {
				job.position += read
				job.progress = float32(job.position) / float32(job.fileSize)
			}
		}
		if err := job.finish(); err != nil {
			log.Error(err)
		}
	}()

	return nil
}

func (job *printJob) nextLine() (int64, error) {

	job.pauseMutex.Lock()
	defer job.pauseMutex.Unlock()

	if job.isCanceled {
		return 0, nil
	}

	line := job.scanner.Text()
	read := int64(len(line) + 1)
	idx := strings.Index(line, ";")
	if idx != -1 {
		line = line[:idx]
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return read, nil
	}

	<-job.manager.printer.QueueGcode(line, false, true)
	return read, nil
}

func (job *printJob) pause() bool {
	if !job.isPaused {
		job.isPaused = true
		job.pauseMutex.Lock()
		job.waitForPrintMoves()
		now := time.Now()
		job.printDuration += now.Sub(job.lastResumeTime)
		job.manager.setState("paused")
		return true
	}
	return false
}

func (job *printJob) resume() bool {
	if job.isPaused {
		job.isPaused = false
		job.pauseMutex.Unlock()
		job.lastResumeTime = time.Now()
		job.manager.setState("printing")
		return true
	}
	return false
}

func (job *printJob) cancel() bool {
	if job.isCanceled {
		return false
	}
	job.isCanceled = true

	if job.isPaused {
		job.isPaused = false
		job.pauseMutex.Unlock()
	}

	job.waitForPrintMoves()
	now := time.Now()
	job.progress = 1
	job.hasEnded = true
	job.endTime = now
	job.lastResumeTime = now
	job.manager.setState("cancelled")
	return true
}

func (job *printJob) finish() error {
	job.waitForPrintMoves()
	now := time.Now()
	job.progress = 1
	job.hasEnded = true
	job.endTime = now
	job.printDuration += now.Sub(job.lastResumeTime)
	job.manager.setState("complete")
	return nil
}

func (job *printJob) waitForPrintMoves() {
	<-job.manager.printer.QueueGcode("M400", false, true)
}

func (job *printJob) getTotalTime() time.Duration {
	if job.hasEnded {
		return job.endTime.Sub(job.startTime)
	} else {
		now := time.Now()
		return now.Sub(job.startTime)
	}
}

func (job *printJob) getPrintTime() time.Duration {
	duration := job.printDuration
	if !job.isPaused && !job.hasEnded {
		now := time.Now()
		duration += now.Sub(job.lastResumeTime)
	}
	return duration
}
