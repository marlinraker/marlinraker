package print_manager

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/files"
	"marlinraker/src/printer_objects"
	"path/filepath"
	"regexp"
	"time"
)

type GcodeDevice interface {
	QueueGcode(gcodeRaw string, important bool, silent bool) (chan string, error)
}

type PrintManager struct {
	printer    GcodeDevice
	state      string
	currentJob *printJob
	ticker     *time.Ticker
	closeCh    chan struct{}
}

var (
	gcodeExtensionRegex = regexp.MustCompile(`(?i)\.gcode$`)
)

func NewPrintManager(printer GcodeDevice) *PrintManager {
	manager := &PrintManager{
		printer: printer,
		state:   "standby",
		ticker:  time.NewTicker(time.Second),
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
				if manager.isPrinting() {
					manager.emit()
				}
			}
		}
	}()
	return manager
}

func (manager *PrintManager) Cleanup() {
	if manager.currentJob != nil {
		manager.currentJob.cancel()
	}
	printer_objects.UnregisterObject("print_stats")
	printer_objects.UnregisterObject("virtual_sdcard")
	printer_objects.UnregisterObject("pause_resume")
	manager.ticker.Stop()
	manager.closeCh <- struct{}{}
	close(manager.closeCh)
}

func (manager *PrintManager) SelectFile(fileName string) error {
	if manager.isPrinting() {
		return errors.New("already printing")
	}
	if !gcodeExtensionRegex.MatchString(fileName) {
		return errors.New("invalid file extension")
	}
	diskPath := filepath.Join(files.DataDir, "gcodes", fileName)
	if _, err := files.Fs.Stat(diskPath); err != nil {
		return err
	}
	manager.currentJob = newPrintJob(manager, fileName)
	manager.emit()
	return nil
}

func (manager *PrintManager) Start() error {
	if !manager.isReadyToPrint() {
		return errors.New("already printing")
	}
	if err := manager.currentJob.start(); err != nil {
		return err
	}
	manager.setState("printing")
	return nil
}

func (manager *PrintManager) Pause() error {
	if manager.currentJob == nil || manager.state != "printing" {
		return errors.New("not currently printing")
	}
	if !manager.currentJob.pause() {
		return errors.New("cannot pause right now")
	}
	return nil
}

func (manager *PrintManager) Resume() error {
	if manager.currentJob == nil || manager.state != "paused" {
		return errors.New("no paused print")
	}
	if !manager.currentJob.resume() {
		return errors.New("cannot resume right now")
	}
	return nil
}

func (manager *PrintManager) Cancel() error {
	if !manager.isPrinting() {
		return errors.New("not currently printing")
	}
	if !manager.currentJob.cancel() {
		return errors.New("print already canceled")
	}
	return nil
}

func (manager *PrintManager) isPrinting() bool {
	if manager.currentJob == nil {
		return false
	}
	return manager.state == "printing" || manager.state == "paused"
}

func (manager *PrintManager) isReadyToPrint() bool {
	if manager.currentJob == nil {
		return false
	}
	switch manager.state {
	case "standby", "complete", "cancelled", "error":
		return true
	}
	return false
}

func (manager *PrintManager) setState(state string) {
	manager.state = state
	manager.emit()
}

func (manager *PrintManager) emit() {
	if err := printer_objects.EmitObject("print_stats"); err != nil {
		log.Error(err)
	}
	if err := printer_objects.EmitObject("virtual_sdcard"); err != nil {
		log.Error(err)
	}
}
