package print_manager

import (
	"bufio"
	"io"
	"marlinraker/src/files"
	"marlinraker/src/printer/parser"
	"marlinraker/src/shared"
	"marlinraker/src/util"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

type printJob struct {
	manager        *PrintManager
	fileName       string
	filePath       string
	pauseCh        chan struct{}
	cancelCh       chan struct{}
	isPaused       atomic.Bool
	isStarted      atomic.Bool
	hasEnded       atomic.Bool
	reader         *bufio.Reader
	position       atomic.Int64
	fileSize       int64
	progress       util.ThreadSafe[float32]
	startTime      util.ThreadSafe[time.Time]
	lastResumeTime util.ThreadSafe[time.Time]
	endTime        util.ThreadSafe[time.Time]
	printDuration  util.ThreadSafe[time.Duration]
}

func newPrintJob(manager *PrintManager, fileName string) *printJob {
	filePath := filepath.Join(files.DataDir, "gcodes", fileName)
	job := &printJob{
		manager:        manager,
		fileName:       fileName,
		filePath:       filePath,
		pauseCh:        make(chan struct{}),
		cancelCh:       make(chan struct{}),
		fileSize:       0,
		progress:       util.NewThreadSafe[float32](0),
		startTime:      util.NewThreadSafe(time.Time{}),
		lastResumeTime: util.NewThreadSafe(time.Time{}),
		endTime:        util.NewThreadSafe(time.Time{}),
		printDuration:  util.NewThreadSafe[time.Duration](0),
	}
	close(job.pauseCh)
	return job
}

func (job *printJob) start(context shared.ExecutorContext) error {

	stat, err := files.Fs.Stat(job.filePath)
	if err != nil {
		return err
	}

	file, err := files.Fs.OpenFile(job.filePath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	now := time.Now()
	job.startTime.Store(now)
	job.lastResumeTime.Store(now)
	job.isStarted.Store(true)
	job.fileSize = stat.Size()
	job.printDuration.Store(0)
	job.position.Store(0)
	job.progress.Store(0)

	reader := bufio.NewReader(file)

	go func() {
		defer func() {
			if err := file.Close(); err != nil {
				util.LogError(err)
			}
		}()

		for {
			bytes, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					util.LogError(err)
				}
				break
			}

			read, line := int64(len(bytes)), strings.TrimRight(string(bytes), "\r\n")
			position := job.position.Add(read)
			job.progress.Store(float32(position) / float32(job.fileSize))

			if canceled, err := job.nextLine(line); err != nil || canceled {
				if err != nil {
					util.LogError(err)
				}
				return
			}
		}

		job.finish("complete", context)
	}()

	return nil
}

func (job *printJob) nextLine(line string) (bool, error) {

	gcode := parser.CleanGcode(line)
	if gcode == "" {
		return false, nil
	}

	context := job.manager.printer.MainExecutorContext()

	select {
	case <-context.Pending():
	case <-job.cancelCh:
		return true, nil
	}

	select {
	case <-job.pauseCh:
	case <-job.cancelCh:
		return true, nil
	}
	<-context.QueueGcode(gcode, true)
	return false, nil
}

func (job *printJob) pause(context shared.ExecutorContext) bool {
	if !job.isPaused.Load() {
		job.isPaused.Store(true)
		job.pauseCh = make(chan struct{})
		job.waitForPrintMoves(context)
		now := time.Now()
		job.printDuration.Do(func(duration time.Duration) time.Duration {
			return duration + now.Sub(job.lastResumeTime.Load())
		})
		job.manager.setState("paused")
		return true
	}
	return false
}

func (job *printJob) resume(context shared.ExecutorContext) bool {
	if job.isPaused.Load() {
		job.waitForPrintMoves(context)
		job.isPaused.Store(false)
		close(job.pauseCh)
		job.lastResumeTime.Store(time.Now())
		job.manager.setState("printing")
		return true
	}
	return false
}

func (job *printJob) cancel(context shared.ExecutorContext) bool {
	if job.hasEnded.Load() || !job.isStarted.Load() {
		return false
	}
	close(job.cancelCh)
	job.finish("cancelled", context)
	return true
}

func (job *printJob) finish(state string, context shared.ExecutorContext) {
	job.waitForPrintMoves(context)
	now := time.Now()
	job.progress.Store(1)
	job.hasEnded.Store(true)
	job.endTime.Store(now)
	if !job.isPaused.Load() {
		job.printDuration.Do(func(duration time.Duration) time.Duration {
			return duration + now.Sub(job.lastResumeTime.Load())
		})
	}
	job.manager.setState(state)
}

func (job *printJob) waitForPrintMoves(context shared.ExecutorContext) {
	<-context.QueueGcode("M400", true)
}

func (job *printJob) getTotalTime() time.Duration {
	if job.hasEnded.Load() {
		return job.endTime.Load().Sub(job.startTime.Load())
	} else {
		now := time.Now()
		return now.Sub(job.startTime.Load())
	}
}

func (job *printJob) getPrintTime() time.Duration {
	duration := job.printDuration.Load()
	if !job.isPaused.Load() && !job.hasEnded.Load() {
		now := time.Now()
		duration += now.Sub(job.lastResumeTime.Load())
	}
	return duration
}
