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
	"marlinraker/src/shared"
	"marlinraker/src/util"
	"strconv"
	"strings"
	"time"
)

type command struct {
	gcode    string
	ch       chan string
	response strings.Builder
}

type watcher interface {
	handle(line string) bool
	stop()
}

type Printer struct {
	config             *config.Config
	context            *executorContext
	path               string
	port               serial.Port
	info               parser.PrinterInfo
	Capabilities       map[string]bool
	IsPrusa            bool
	hasEmergencyParser bool
	limits             parser.PrinterLimits
	watchers           util.ThreadSafe[[]watcher]
	CloseCh            chan struct{}
	connected          bool
	heaters            heatersObject
	PrintManager       *print_manager.PrintManager
	MacroManager       *macros.MacroManager
	GcodeState         *GcodeState
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
			Position:             [4]float32{0, 0, 0, 0},
			IsAbsoluteCoordinate: true,
			IsAbsoluteExtrude:    true,
			SpeedFactor:          100,
			ExtrudeFactor:        100,
			Feedrate:             0,
		},
		savedGcodeStates: make(map[string]GcodeState),
	}
	printer.context = newExecutorContext(printer, "main")
	printer.PrintManager = print_manager.NewPrintManager(printer)
	printer.MacroManager = macros.NewMacroManager(printer, printer.config)

	go printer.readPort()
	err = printer.tryToConnect()
	if err != nil {
		return nil, err
	}

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

func (printer *Printer) tryToConnect() error {
	maxAttempts := printer.config.Serial.MaxConnectionAttempts
	for i := 0; i < maxAttempts; i++ {
		if err := printer.handshake(); err != nil {
			if i < maxAttempts-1 {
				log.Error(err)
				log.Println("Retrying... (" + strconv.Itoa(maxAttempts-i-1) + " attempt(s) left)")
			}
			continue
		}
		return printer.setup()
	}
	return errors.New("could not connect to printer after " + strconv.Itoa(maxAttempts) + " attempts")
}

func (printer *Printer) handshake() error {
	printer.context.resetCommandQueue()

	errCh1, errCh2 := make(chan error), make(chan error)

	go func() {
		defer close(errCh1)

		for {
			info, capabilities, err := parser.ParseM115(<-printer.context.QueueGcode("M115", false, true))
			if err != nil {
				log.Error(err)
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
	printer.cleanup()
	close(printer.CloseCh)
}

func (printer *Printer) cleanup() {
	for _, watcher := range printer.watchers.Get() {
		watcher.stop()
	}
	printer.PrintManager.Cleanup(printer.context)
	printer.MacroManager.Cleanup()
}

func (printer *Printer) setup() error {

	errorCh1, errorCh2 := make(chan error), make(chan error)

	go func() {
		defer close(errorCh1)

		printer.IsPrusa = strings.HasPrefix(printer.info.FirmwareName, "Prusa-Firmware")
		if printer.IsPrusa {
			log.Println("Printer runs Prusa-Firmware")
		}

		printer.hasEmergencyParser = printer.IsPrusa || printer.Capabilities["EMERGENCY_PARSER"]

		reportVelocity := printer.config.Misc.ReportVelocity

		if printer.IsPrusa {
			c := 1 << 0
			if !reportVelocity {
				c |= 1 << 2
			}
			<-printer.context.QueueGcode("M155 S1 C"+strconv.Itoa(c), false, true)
		}

		for {
			ch := printer.context.QueueGcode("M503", false, true)
			limits, err := parser.ParseM503(<-ch)
			if err != nil {
				log.Println(err)
				continue
			}
			printer.limits = limits
			break
		}

		tempWatcher := newTempWatcher(printer)
		printer.watchers.Set(append(printer.watchers.Get(), tempWatcher))
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
	if err := printer.GcodeState.update(line); err != nil {
		log.Error(err)
	}
}

func (printer *Printer) handleResponseLine(line string) bool {

	for _, watcher := range printer.watchers.Get() {
		if watcher.handle(line) {
			return true
		}
	}

	if printer.connected && strings.HasPrefix(line, "echo:") {
		message := line[5:]
		if strings.HasPrefix(message, "busy:") {
			return false
		}

		err := notification.Publish(notification.New("notify_gcode_response", []any{message}))
		if err != nil {
			log.Error(err)
		}

		gcode_store.LogNow(message, gcode_store.Response)
		return true
	}

	return false
}

func (printer *Printer) checkEmergencyCommand(gcode string) bool {
	if printer.hasEmergencyParser && parser.IsEmergencyCommand(gcode, printer.IsPrusa) {
		log.Debugln("emergency: " + gcode)
		_, _ = printer.port.Write([]byte(gcode + "\n"))
		printer.handleRequestLine(gcode)
		return true
	}
	return false
}

func (printer *Printer) readLine(line string) {
	if printer.handleResponseLine(line) {
		return
	}
	printer.context.readLine(line)
}

func (printer *Printer) GetPrintManager() shared.PrintManager {
	return printer.PrintManager
}

func (printer *Printer) Respond(message string) error {
	gcode_store.LogNow(message, gcode_store.Response)
	return notification.Publish(notification.New("notify_gcode_response", []any{message}))
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
