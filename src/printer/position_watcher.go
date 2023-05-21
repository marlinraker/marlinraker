package printer

import (
	"marlinraker/src/printer/parser"
	"marlinraker/src/printer_objects"
	"marlinraker/src/util"
	"math"
	"strings"
	"time"
)

type positionWatcher struct {
	printer        *Printer
	ticker         *time.Ticker
	closeCh        chan struct{}
	last           time.Time
	autoReport     bool
	reportVelocity bool
}

func newPositionWatcher(printer *Printer) *positionWatcher {

	watcher := &positionWatcher{
		printer:        printer,
		closeCh:        make(chan struct{}),
		reportVelocity: printer.config.Printer.Gcode.ReportVelocity,
		autoReport: !printer.config.Printer.Gcode.ReportVelocity &&
			(printer.Capabilities["AUTOREPORT_POS"] || printer.Capabilities["AUTOREPORT_POSITION"]),
	}

	if watcher.autoReport && !printer.IsPrusa {
		<-printer.context.QueueGcode("M154 S1", true)
	} else if !watcher.autoReport {
		go watcher.runTimer()
	}

	return watcher
}

func (watcher *positionWatcher) handle(line string) {
	if !watcher.autoReport || !strings.HasPrefix(line, "X:") {
		return
	}
	watcher.readPos(line)
}

func (watcher *positionWatcher) readPos(line string) {
	position, err := parser.ParseM114(line)
	if err != nil {
		util.LogError(err)
		return
	}

	if watcher.reportVelocity {
		now := time.Now()
		if !watcher.last.IsZero() {
			dt := now.Sub(watcher.last).Seconds()
			if dt < 0.5 {
				oldPos := watcher.printer.GcodeState.Position
				dxy := math.Hypot(position[0]-oldPos[0], position[1]-oldPos[1])
				watcher.printer.GcodeState.Velocity = dxy / dt
				de := position[3] - oldPos[3]
				watcher.printer.GcodeState.EVelocity = de / dt
			} else {
				watcher.printer.GcodeState.Velocity = 0
				watcher.printer.GcodeState.EVelocity = 0
			}
		}
		watcher.last = now
	}

	for i := 0; i < 4; i++ {
		watcher.printer.GcodeState.Position[i] = position[i]
	}

	if err := printer_objects.EmitObject("toolhead", "motion_report", "gcode_move"); err != nil {
		util.LogError(err)
	}
}

func (watcher *positionWatcher) stop() {
	if watcher.ticker != nil {
		watcher.ticker.Stop()
	}
	close(watcher.closeCh)
}

func (watcher *positionWatcher) runTimer() {

	duration := time.Second
	if watcher.reportVelocity {
		duration = 200 * time.Millisecond
	}

	watcher.ticker = time.NewTicker(duration)
	for {
		select {
		case <-watcher.closeCh:
			break
		case <-watcher.ticker.C:
			watcher.tick()
		}
	}
}

func (watcher *positionWatcher) tick() {
	if !watcher.reportVelocity && watcher.printer.PrintManager.GetState() == "printing" {
		// prevent stuttering while printing
		return
	}
	response := <-watcher.printer.context.QueueGcode("M114 R", true)
	watcher.readPos(response)
}
