package print_manager

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/files"
	"marlinraker/src/printer_objects"
	"marlinraker/src/shared"
	"marlinraker/src/util"
	"path/filepath"
	"regexp"
	"sync/atomic"
	"time"
)

type PrintManager struct {
	printer    shared.Printer
	state      util.ThreadSafe[string]
	currentJob atomic.Pointer[printJob]
	ticker     *time.Ticker
	closeCh    chan struct{}
}

var (
	gcodeExtensionRegex = regexp.MustCompile(`(?i)\.gcode$`)
)

func NewPrintManager(printer shared.Printer) *PrintManager {
	manager := &PrintManager{
		printer: printer,
		state:   util.NewThreadSafe("standby"),
		ticker:  time.NewTicker(time.Second),
		closeCh: make(chan struct{}),
	}
	printer_objects.RegisterObject("print_stats", printStatsObject{manager})
	printer_objects.RegisterObject("virtual_sdcard", virtualSdcardObject{manager})
	printer_objects.RegisterObject("pause_resume", pauseResumeObject{manager})
	go func() {
		for {
			select {
			case <-manager.closeCh:
				return
			case <-manager.ticker.C:
				if manager.isPrinting(manager.currentJob.Load(), manager.state.Load()) {
					if err := manager.emit(); err != nil {
						log.Error(err)
					}
				}
			}
		}
	}()
	return manager
}

func (manager *PrintManager) Cleanup(context shared.ExecutorContext) {
	if job := manager.currentJob.Load(); job != nil {
		job.cancel(context)
	}
	printer_objects.UnregisterObject("print_stats")
	printer_objects.UnregisterObject("virtual_sdcard")
	printer_objects.UnregisterObject("pause_resume")
	manager.ticker.Stop()
	manager.closeCh <- struct{}{}
	close(manager.closeCh)
}

func (manager *PrintManager) SelectFile(fileName string) error {
	job, state := manager.currentJob.Load(), manager.state.Load()
	if manager.isPrinting(job, state) {
		return errors.New("already printing")
	}
	if !gcodeExtensionRegex.MatchString(fileName) {
		return errors.New("invalid file extension")
	}
	diskPath := filepath.Join(files.DataDir, "gcodes", fileName)
	if _, err := files.Fs.Stat(diskPath); err != nil {
		return err
	}
	manager.currentJob.Store(newPrintJob(manager, fileName))
	if err := manager.emit(); err != nil {
		log.Error(err)
	}
	return nil
}

func (manager *PrintManager) Start(context shared.ExecutorContext) error {
	job, state := manager.currentJob.Load(), manager.state.Load()
	if !manager.isReadyToPrint(job, state) {
		return errors.New("failed to start print: already printing")
	}
	if err := job.start(context); err != nil {
		return fmt.Errorf("failed to start print: %w", err)
	}
	if err := manager.setState("printing"); err != nil {
		return fmt.Errorf("failed to start print: %w", err)
	}
	return nil
}

func (manager *PrintManager) Pause(context shared.ExecutorContext) error {
	job, state := manager.currentJob.Load(), manager.state.Load()
	if job == nil || state != "printing" {
		return errors.New("failed to pause print: not currently printing")
	}
	if !job.pause(context) {
		return errors.New("failed to pause print: cannot pause right now")
	}
	return nil
}

func (manager *PrintManager) Resume(context shared.ExecutorContext) error {
	job, state := manager.currentJob.Load(), manager.state.Load()
	if job == nil || state != "paused" {
		return errors.New("failed to resume print: no paused print")
	}
	if !job.resume(context) {
		return errors.New("failed to resume print: cannot resume right now")
	}
	return nil
}

func (manager *PrintManager) Cancel(context shared.ExecutorContext) error {
	job, state := manager.currentJob.Load(), manager.state.Load()
	if !manager.isPrinting(job, state) {
		return errors.New("not currently printing")
	}
	if !job.cancel(context) {
		return errors.New("print already canceled")
	}
	return nil
}

func (manager *PrintManager) Reset(context shared.ExecutorContext) error {
	job, state := manager.currentJob.Load(), manager.state.Load()
	if job == nil {
		return errors.New("no file selected")
	}
	if manager.isPrinting(job, state) {
		if err := manager.Cancel(context); err != nil {
			return fmt.Errorf("failed to reset print: %w", err)
		}
	}
	manager.currentJob.Store(nil)
	if err := manager.setState("standby"); err != nil {
		return fmt.Errorf("failed to reset print: %w", err)
	}
	return nil
}

func (manager *PrintManager) GetState() string {
	return manager.state.Load()
}

func (manager *PrintManager) CanPrint(fileName string) bool {
	job, state := manager.currentJob.Load(), manager.state.Load()
	if manager.isPrinting(job, state) || !gcodeExtensionRegex.MatchString(fileName) {
		return false
	}
	if _, err := files.Fs.Stat(filepath.Join(files.DataDir, "gcodes", fileName)); err != nil {
		return false
	}
	return true
}

func (manager *PrintManager) IsPrinting() bool {
	return manager.isPrinting(manager.currentJob.Load(), manager.state.Load())
}

func (manager *PrintManager) isPrinting(job *printJob, state string) bool {
	if job == nil {
		return false
	}
	return state == "printing" || state == "paused"
}

func (manager *PrintManager) isReadyToPrint(job *printJob, state string) bool {
	if job == nil {
		return false
	}
	switch state {
	case "standby", "complete", "cancelled", "error":
		return true
	}
	return false
}

func (manager *PrintManager) setState(state string) error {
	manager.state.Store(state)
	if err := manager.emit(); err != nil {
		return fmt.Errorf("failed set printer state to %q: %w", state, err)
	}
	return nil
}

func (manager *PrintManager) emit() error {
	if err := printer_objects.EmitObject("print_stats", "virtual_sdcard"); err != nil {
		return fmt.Errorf("failed to emit print stats: %w", err)
	}
	return nil
}

func (manager *PrintManager) getFilamentUsed() float64 {
	if job := manager.currentJob.Load(); job != nil {
		return job.getFilamentUsed()
	}
	return 0
}
