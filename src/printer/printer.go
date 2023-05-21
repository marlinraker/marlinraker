package printer

import (
	"bufio"
	"errors"
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
	"strconv"
	"strings"
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
	Error              error
	config             *config.Config
	context            *executorContext
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
		return nil, err
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
		return nil, err
	}

	printer_objects.RegisterObject("toolhead", toolheadObject{printer})
	printer_objects.RegisterObject("motion_report", motionReportObject{printer})
	printer_objects.RegisterObject("gcode_move", gcodeMoveObject{printer})

	printer.connected = true
	return printer, nil
}

func (printer *Printer) MainExecutorContext() shared.ExecutorContext {
	return printer.context
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
	<-printer.context.QueueGcode("M112", true)
}

func (printer *Printer) tryToConnect() error {
	maxAttempts := printer.config.Serial.MaxConnectionAttempts
	for i := 0; i < maxAttempts; i++ {
		if err := printer.handshake(i); err != nil {
			if i < maxAttempts-1 {
				util.LogError(err)
				log.Println("Retrying... (" + strconv.Itoa(maxAttempts-i-1) + " attempt(s) left)")
			}
			continue
		}
		return printer.setup()
	}
	return errors.New("could not connect to printer after " + strconv.Itoa(maxAttempts) + " attempts")
}

func (printer *Printer) handshake(attempt int) error {
	printer.context = newExecutorContext(printer, "handshake"+strconv.Itoa(attempt))
	defer printer.context.close()

	errCh1, errCh2 := make(chan error), make(chan error)

	go func() {
		defer close(errCh1)

		for {
			info, capabilities, err := parser.ParseM115(<-printer.context.QueueGcode("M115", true))
			if err != nil {
				util.LogError(err)
				continue
			}

			printer.info, printer.Capabilities = info, capabilities
			break
		}

		log.Println("Identified " + printer.info.MachineType + " on " + printer.info.FirmwareName + " with " + strconv.Itoa(len(printer.Capabilities)) + " capabilities")
		errCh1 <- nil
	}()

	go func() {
		defer close(errCh2)
		time.Sleep(time.Duration(printer.config.Serial.ConnectionTimeout) * time.Millisecond)
		errCh2 <- errors.New("printer handshake took too long")
	}()

	var err error
	select {
	case err = <-errCh1:
	case err = <-errCh2:
	}
	return err
}

func (printer *Printer) readPort() {
	scanner := bufio.NewScanner(printer.port)
	for scanner.Scan() {
		line := scanner.Text()
		log.WithField("port", printer.path).Debugln("recv: " + line)
		printer.readLine(line)
	}
	if err := scanner.Err(); err != nil {
		if err, isPortError := err.(*serial.PortError); isPortError && err.Code() == serial.PortClosed {
			log.Println("Port " + printer.path + " has been closed")
		} else {
			util.LogError(err)
		}
	}
	printer.cleanup()
	close(printer.CloseCh)
}

func (printer *Printer) cleanup() {
	for _, watcher := range printer.watchers.Load() {
		watcher.stop()
	}
	printer.PrintManager.Cleanup(printer.context)
	printer.MacroManager.Cleanup()
	printer_objects.UnregisterObject("toolhead")
	printer_objects.UnregisterObject("motion_report")
	printer_objects.UnregisterObject("gcode_move")
}

func (printer *Printer) setup() error {

	printer.context = newExecutorContext(printer, "main")
	errorCh1, errorCh2 := make(chan error), make(chan error)

	go func() {
		defer close(errorCh1)

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
			<-printer.context.QueueGcode("M155 S1 C"+strconv.Itoa(c), true)
		}

		for {
			ch := printer.context.QueueGcode("M503", true)
			limits, err := parser.ParseM503(<-ch)
			if err != nil {
				log.Println(err)
				continue
			}
			printer.limits = limits
			break
		}

		tempWatcher, positionWatcher := newTempWatcher(printer), newPositionWatcher(printer)
		printer.watchers.Do(func(watchers []watcher) []watcher {
			return append(watchers, tempWatcher, positionWatcher)
		})
		printer.heaters = <-tempWatcher.heatersCh

		errorCh1 <- nil
	}()

	go func() {
		defer close(errorCh2)
		time.Sleep(10 * time.Second)
		errorCh2 <- errors.New("printer initialization took too long")
	}()

	var err error
	select {
	case err = <-errorCh1:
	case err = <-errorCh2:
	}

	return err
}

func (printer *Printer) handleRequestLine(line string) {
	switch {
	case parser.M112.MatchString(line):
		printer.Error = errors.New("emergency stop")
		if err := printer.Disconnect(); err != nil {
			util.LogError(err)
		}

	default:
		if err := printer.GcodeState.update(line); err != nil {
			util.LogError(err)
		}
	}
}

func (printer *Printer) handleResponseLine(line string) bool {

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
			util.LogError(err)
		}

		gcode_store.LogNow(message, gcode_store.Response)
		return true
	}

	return false
}

func (printer *Printer) executeEmergencyCommand(gcode string) bool {
	if printer.hasEmergencyParser && parser.IsEmergencyCommand(gcode) {
		log.Debugln("emergency: " + gcode)
		if _, err := printer.port.Write([]byte(gcode + "\n")); err != nil {
			util.LogError(err)
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
	if printer.context != nil {
		printer.context.readLine(line)
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
		return errors.New("there is no saved G-code state with the name \"" + name + "\"")
	}
	delete(printer.savedGcodeStates, name)
	printer.GcodeState.restore(context, savedState)
	return nil
}
