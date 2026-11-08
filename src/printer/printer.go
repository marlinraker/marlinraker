package printer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
	"marlinraker/src/api/notification"
	"marlinraker/src/config"
	"marlinraker/src/marlinraker/gcode_store"
	"marlinraker/src/printer/macros"
	"marlinraker/src/printer/parser"
	"marlinraker/src/printer/print_manager"
	"marlinraker/src/printer_objects"
	"marlinraker/src/shared"
	"marlinraker/src/util"
	"strings"
	"sync/atomic"
	"time"
)

type watcher interface {
	handle(line string)
	stop()
}

type Printer struct {
	Capabilities       map[string]bool
	IsPrusa            bool
	CloseCh            chan struct{}
	PrintManager       *print_manager.PrintManager
	MacroManager       *macros.MacroManager
	GcodeState         *GcodeState
	err                atomic.Pointer[error]
	config             *config.Config
	context            atomic.Pointer[executorContext]
	path               string
	port               serial.Port
	info               parser.PrinterInfo
	hasEmergencyParser bool
	limits             parser.PrinterLimits
	watchers           util.ThreadSafe[[]watcher]
	connected          bool
	heaters            heatersObject
	savedGcodeStates   map[string]GcodeState
}

func New(config *config.Config, path string, baudRate int) (*Printer, error) {

	port, err := serial.Open(path, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %q: %w", path, err)
	}

	printer := &Printer{
		config:    config,
		path:      path,
		port:      port,
		watchers:  util.NewThreadSafe(make([]watcher, 0)),
		CloseCh:   make(chan struct{}),
		connected: false,
		GcodeState: &GcodeState{
			Position:             [4]float64{0, 0, 0, 0},
			IsAbsoluteCoordinate: true,
			IsAbsoluteExtrude:    true,
			SpeedFactor:          100,
			ExtrudeFactor:        100,
			Feedrate:             0,
		},
		savedGcodeStates: make(map[string]GcodeState),
	}
	printer.PrintManager = print_manager.NewPrintManager(printer)
	printer.MacroManager = macros.NewMacroManager(printer, printer.config)

	go printer.readPort()
	err = printer.tryToConnect()
	if err != nil {
		// Release the serial port so that the next connection attempt can open it again.
		// Closing the port also stops readPort, which runs the cleanup.
		if closeErr := printer.Disconnect(); closeErr != nil {
			log.Errorf("Failed to close serial port %q: %v", path, closeErr)
		}
		return nil, err
	}

	printer_objects.RegisterObject("toolhead", toolheadObject{printer})
	printer_objects.RegisterObject("motion_report", motionReportObject{printer})
	printer_objects.RegisterObject("gcode_move", gcodeMoveObject{printer})

	printer.connected = true
	return printer, nil
}

// GetError returns the error that caused the printer to disconnect, if any.
func (printer *Printer) GetError() error {
	if err := printer.err.Load(); err != nil {
		return *err
	}
	return nil
}

func (printer *Printer) setError(err error) {
	printer.err.Store(&err)
}

func (printer *Printer) MainExecutorContext() shared.ExecutorContext {
	return printer.ctx()
}

// ctx returns the executor context that currently owns the serial connection.
func (printer *Printer) ctx() *executorContext {
	return printer.context.Load()
}

func (printer *Printer) Disconnect() error {
	err := printer.port.Close()
	if err != nil {
		return err
	}
	<-printer.CloseCh
	return nil
}

func (printer *Printer) EmergencyStop() {
	<-printer.ctx().QueueGcode("M112", true)
}

func (printer *Printer) tryToConnect() error {
	maxAttempts := printer.config.Serial.MaxConnectionAttempts
	for i := 0; i < maxAttempts; i++ {
		if err := printer.handshake(i); err != nil {
			if i < maxAttempts-1 {
				log.Errorf("Printer handshake failed: %v", err)
				log.Printf("Retrying... (%d attempt(s) left)", maxAttempts-i-1)
			}
			continue
		}
		return printer.setup()
	}
	return fmt.Errorf("could not connect to printer after %d attempts", maxAttempts)
}

func (printer *Printer) handshake(attempt int) error {
	printer.context.Store(newExecutorContext(printer, fmt.Sprintf("handshake%d", attempt)))
	defer printer.ctx().close()

	timeout := time.Duration(printer.config.Serial.ConnectionTimeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		response, err := printer.ctx().await(ctx, printer.ctx().QueueGcode("M115", true))
		if err != nil {
			return errors.New("printer handshake took too long")
		}

		info, capabilities, err := parser.ParseM115(response)
		if err != nil {
			log.Errorf("Failed to parse capabilities: %v", err)
			continue
		}

		printer.info, printer.Capabilities = info, capabilities
		log.Printf("Identified %s on %s with %d capabilities", printer.info.MachineType, printer.info.FirmwareName, len(printer.Capabilities))
		return nil
	}
}

func (printer *Printer) readPort() {
	scanner := bufio.NewScanner(printer.port)
	for scanner.Scan() {
		line := scanner.Text()
		log.WithField("port", printer.path).Debugf("recv: %s", line)
		printer.readLine(line)
	}
	if err := scanner.Err(); err != nil {
		var portErr *serial.PortError
		if errors.As(err, &portErr) && portErr.Code() == serial.PortClosed {
			log.Printf("Port %s has been closed", printer.path)
		} else {
			log.Errorf("Failed to read from port %q: %v", printer.path, err)
		}
	}
	printer.cleanup()
	close(printer.CloseCh)
}

func (printer *Printer) cleanup() {
	// The port is closed at this point, so nothing will answer pending commands anymore.
	// Closing the context releases everyone waiting for a response.
	if ctx := printer.ctx(); ctx != nil {
		ctx.close()
	}
	for _, watcher := range printer.watchers.Load() {
		watcher.stop()
	}
	printer.PrintManager.Cleanup(printer.ctx())
	printer.MacroManager.Cleanup()
	printer_objects.UnregisterObject("toolhead")
	printer_objects.UnregisterObject("motion_report")
	printer_objects.UnregisterObject("gcode_move")
}

func (printer *Printer) setup() error {

	printer.context.Store(newExecutorContext(printer, "main"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	timeoutErr := errors.New("printer initialization took too long")

	printer.IsPrusa = strings.HasPrefix(printer.info.FirmwareName, "Prusa-Firmware")
	if printer.IsPrusa {
		log.Println("Printer runs Prusa-Firmware")
	}

	printer.hasEmergencyParser = printer.IsPrusa || printer.Capabilities["EMERGENCY_PARSER"]

	if printer.IsPrusa {
		c := 1 << 0
		if !printer.config.Printer.Gcode.ReportVelocity {
			c |= 1 << 2
		}
		if _, err := printer.ctx().await(ctx, printer.ctx().QueueGcode(fmt.Sprintf("M155 S1 C%d", c), true)); err != nil {
			return timeoutErr
		}
	}

	for {
		response, err := printer.ctx().await(ctx, printer.ctx().QueueGcode("M503", true))
		if err != nil {
			return timeoutErr
		}
		limits, err := parser.ParseM503(response)
		if err != nil {
			log.Errorf("Failed parsing limits: %v", err)
			continue
		}
		printer.limits = limits
		break
	}

	tempWatcher, positionWatcher := newTempWatcher(printer), newPositionWatcher(printer)
	printer.watchers.Do(func(watchers []watcher) []watcher {
		return append(watchers, tempWatcher, positionWatcher)
	})

	select {
	case printer.heaters = <-tempWatcher.heatersCh:
		return nil
	case <-ctx.Done():
		return timeoutErr
	}
}

func (printer *Printer) handleRequestLine(line string) {
	switch {
	case parser.M112.MatchString(line):
		printer.setError(errors.New("emergency stop"))
		if err := printer.Disconnect(); err != nil {
			log.Errorf("Failed updating state: %v", err)
		}

	default:
		if err := printer.GcodeState.update(line); err != nil {
			log.Errorf("Failed updating state: %v", err)
		}
	}
}

func (printer *Printer) handleResponseLine(line string) bool {

	if strings.HasPrefix(line, "//") {
		action := parser.ParseAction(line)
		switch action {
		case "cancel":
			log.Println("Canceling print")
			_ = printer.ctx().QueueGcode("CANCEL_PRINT", true)
		case "pause":
			log.Println("Pausing print")
			_ = printer.ctx().QueueGcode("PAUSE", true)
		case "resume":
			log.Println("Resuming print")
			_ = printer.ctx().QueueGcode("RESUME", true)
		}
		return true
	}

	for _, watcher := range printer.watchers.Load() {
		watcher.handle(line)
	}

	if printer.connected && strings.HasPrefix(line, "echo:") {
		message := line[5:]
		if strings.HasPrefix(message, "busy:") {
			return false
		}

		err := notification.Publish(notification.New("notify_gcode_response", []any{message}))
		if err != nil {
			log.Errorf("Error publishing notification: %v", err)
		}

		gcode_store.LogNow(message, gcode_store.Response)
		return true
	}

	return false
}

func (printer *Printer) executeEmergencyCommand(gcode string) bool {
	if printer.hasEmergencyParser && parser.IsEmergencyCommand(gcode) {
		log.Debugf("emergency: %s", gcode)
		if _, err := fmt.Fprintln(printer.port, gcode); err != nil {
			log.Errorf("Failed writing to printer port: %v", err)
		}
		printer.handleRequestLine(gcode)
		return true
	}
	return false
}

func (printer *Printer) readLine(line string) {
	if printer.handleResponseLine(line) {
		return
	}
	if ctx := printer.ctx(); ctx != nil {
		ctx.readLine(line)
	}
}

func (printer *Printer) GetPrintManager() shared.PrintManager {
	return printer.PrintManager
}

func (printer *Printer) Respond(message string) error {
	gcode_store.LogNow(message, gcode_store.Response)
	return notification.Publish(notification.New("notify_gcode_response", []any{message}))
}

func (printer *Printer) GetGcodeState() shared.GcodeState {
	return printer.GcodeState
}

func (printer *Printer) SaveGcodeState(name string) {
	currentState := *printer.GcodeState
	printer.savedGcodeStates[name] = currentState
}

func (printer *Printer) RestoreGcodeState(context shared.ExecutorContext, name string) error {
	savedState, exists := printer.savedGcodeStates[name]
	if !exists {
		return fmt.Errorf("there is no saved G-code state with the name %q", name)
	}
	delete(printer.savedGcodeStates, name)
	printer.GcodeState.restore(context, savedState)
	return nil
}
