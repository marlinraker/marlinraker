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
	"strconv"
	"strings"
	"sync"
	"time"
)

type Command struct {
	gcode       string
	ch          chan string
	ResponseBuf string
}

type watcher interface {
	handle(line string) bool
	stop()
}

type Printer struct {
	config                 *config.Config
	path                   string
	port                   serial.Port
	ready                  bool
	commandQueue           []Command
	commandQueueMutex      *sync.RWMutex
	currentCommand         *Command
	currentCommandMutex    *sync.RWMutex
	info                   parser.PrinterInfo
	Capabilities           map[string]bool
	IsPrusa                bool
	hasEmergencyParser     bool
	limits                 parser.PrinterLimits
	IsAbsolutePositioning  bool
	IsAbsoluteEPositioning bool
	watchers               []watcher
	watchersMutex          *sync.RWMutex
	CloseCh                chan struct{}
	connected              bool
	heaters                heatersObject
	PrintManager           *print_manager.PrintManager
	macroManager           *macros.MacroManager
}

func New(config *config.Config, path string, baudRate int) (*Printer, error) {

	port, err := serial.Open(path, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return nil, err
	}

	printer := &Printer{
		config:                 config,
		path:                   path,
		port:                   port,
		ready:                  true,
		IsAbsolutePositioning:  true,
		IsAbsoluteEPositioning: true,
		CloseCh:                make(chan struct{}),
		connected:              false,
		commandQueueMutex:      &sync.RWMutex{},
		watchersMutex:          &sync.RWMutex{},
		currentCommandMutex:    &sync.RWMutex{},
	}
	printer.PrintManager = print_manager.NewPrintManager(printer)
	printer.macroManager = macros.NewMacroManager(printer)

	go printer.readPort()
	err = printer.tryToConnect()
	if err != nil {
		return nil, err
	}

	printer.connected = true
	return printer, nil
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
	printer.resetCommandQueue()

	errCh1, errCh2 := make(chan error), make(chan error)

	go func() {
		defer close(errCh1)

		for {
			info, capabilities, err := parser.ParseM115(<-printer.QueueGcode("M115", false, true))
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
	printer.watchersMutex.RLock()
	for _, watcher := range printer.watchers {
		watcher.stop()
	}
	printer.watchersMutex.RUnlock()
	printer.PrintManager.Cleanup()
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
			<-printer.QueueGcode("M155 S1 C"+strconv.Itoa(c), false, true)
		}

		for {
			ch := printer.QueueGcode("M503", false, true)
			limits, err := parser.ParseM503(<-ch)
			if err != nil {
				log.Println(err)
				continue
			}
			printer.limits = limits
			break
		}

		tempWatcher := newTempWatcher(printer)
		printer.watchersMutex.Lock()
		printer.watchers = append(printer.watchers, tempWatcher)
		printer.watchersMutex.Unlock()
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

	case parser.G90.MatchString(line):
		printer.IsAbsolutePositioning = true
		printer.IsAbsoluteEPositioning = true

	case parser.G91.MatchString(line):
		printer.IsAbsolutePositioning = false
		printer.IsAbsoluteEPositioning = false
	}
}

func (printer *Printer) handleResponseLine(line string) bool {

	printer.watchersMutex.RLock()
	for _, watcher := range printer.watchers {
		if watcher.handle(line) {
			return false
		}
	}
	printer.watchersMutex.RUnlock()

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
		return false
	}

	return true
}

func (printer *Printer) QueueGcode(gcodeRaw string, important bool, silent bool) chan string {

	if !silent {
		gcode_store.LogNow(gcodeRaw, gcode_store.Command)
	}

	if strings.Contains(gcodeRaw, "\n") {
		ch := make(chan string)
		go func() {
			chans := make([]chan string, 0)
			for _, line := range strings.Split(gcodeRaw, "\n") {
				chans = append(chans, printer.QueueGcode(line, important, true))
			}
			responses := make([]string, 0)
			for _, responseCh := range chans {
				responses = append(responses, <-responseCh)
			}
			ch <- strings.Join(responses, "\n")
		}()
		return ch
	}

	gcode := strings.TrimSpace(strings.Split(gcodeRaw, ";")[0])
	if gcode == "" {
		return nil
	}

	if errCh := printer.macroManager.TryCommand(gcode); errCh != nil {
		ch := make(chan string)
		go func() {
			if err := <-errCh; err != nil {
				message := "!! Error: " + err.Error()
				gcode_store.LogNow(message, gcode_store.Response)
				err = notification.Publish(notification.New("notify_gcode_response", []any{message}))
				if err != nil {
					log.Error(err)
				}
			}
			ch <- "ok"
		}()
		return ch
	}

	if printer.hasEmergencyParser && parser.IsEmergencyCommand(gcode, printer.IsPrusa) {
		log.Debugln("emergency: " + gcode)
		_, _ = printer.port.Write([]byte(gcode + "\n"))
		printer.handleRequestLine(gcode)
		printer.flush()
		return nil

	} else {
		ch := make(chan string)
		printer.commandQueueMutex.Lock()
		printer.commandQueue = append(printer.commandQueue, Command{gcode: gcode, ch: ch})
		printer.commandQueueMutex.Unlock()
		printer.flush()
		return ch
	}
}

func (printer *Printer) readLine(line string) {

	printer.currentCommandMutex.RLock()
	currentCommand := printer.currentCommand
	printer.currentCommandMutex.RUnlock()

	if printer.handleResponseLine(line) {
		if currentCommand != nil {
			if currentCommand.ResponseBuf == "" {
				currentCommand.ResponseBuf = line
			} else {
				currentCommand.ResponseBuf += "\n" + line
			}
		}
	}

	if strings.HasPrefix(line, "ok") {
		if currentCommand != nil {
			currentCommand.ch <- currentCommand.ResponseBuf
			close(currentCommand.ch)
		}
		printer.currentCommand = nil
		printer.ready = true
		printer.flush()
		return
	}
}

func (printer *Printer) flush() {

	printer.commandQueueMutex.Lock()
	defer printer.commandQueueMutex.Unlock()

	if !printer.ready || len(printer.commandQueue) == 0 {
		return
	}

	printer.currentCommandMutex.Lock()
	printer.currentCommand = &printer.commandQueue[0]
	printer.currentCommandMutex.Unlock()

	printer.commandQueue = printer.commandQueue[1:]

	log.WithField("port", printer.path).Debugln("write: " + string(printer.currentCommand.gcode))

	_, err := printer.port.Write([]byte(printer.currentCommand.gcode + "\n"))
	if err != nil {
		log.Error(err)
	}
	printer.handleRequestLine(printer.currentCommand.gcode)

	printer.ready = false
}

func (printer *Printer) resetCommandQueue() {
	printer.commandQueueMutex.Lock()
	printer.currentCommandMutex.Lock()
	defer printer.commandQueueMutex.Unlock()
	defer printer.currentCommandMutex.Unlock()

	printer.commandQueue = []Command{}
	printer.currentCommand = nil
	printer.ready = true
}
