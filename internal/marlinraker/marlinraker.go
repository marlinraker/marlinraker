package marlinraker

import (
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/api/notification"
	"github.com/marlinraker/marlinraker/internal/config"
	"github.com/marlinraker/marlinraker/internal/constants"
	"github.com/marlinraker/marlinraker/internal/marlinraker/temp_store"
	"github.com/marlinraker/marlinraker/internal/printer"
	"github.com/marlinraker/marlinraker/internal/printer_objects"
	"github.com/marlinraker/marlinraker/internal/scanner"
	"github.com/marlinraker/marlinraker/internal/system_info"
)

type KlippyState string

const (
	Ready    KlippyState = "ready"
	Error    KlippyState = "error"
	Shutdown KlippyState = "shutdown"
	Startup  KlippyState = "startup"
)

var (
	Config          *config.Config
	KlipperSettings map[string]any
	KlipperConfig   map[string]any

	stateMutex   sync.RWMutex
	state        = Shutdown
	stateMessage string

	currentPrinter atomic.Pointer[printer.Printer]
	// connectMutex is held for the whole lifetime of a connection attempt and
	// the resulting printer connection.
	connectMutex sync.Mutex
)

// GetState returns the current Klippy state and its message.
func GetState() (KlippyState, string) {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	return state, stateMessage
}

// GetPrinter returns the connected printer or nil if no printer is connected.
func GetPrinter() *printer.Printer {
	return currentPrinter.Load()
}

func Init(cfg *config.Config) {
	log.Printf("Starting Marlinraker %s", constants.Version)

	Config = cfg
	KlipperSettings, KlipperConfig = config.GenerateFakeKlipperConfig(cfg)

	if cfg.Misc.ExtendedLogs {
		log.SetLevel(log.DebugLevel)
	}

	printer_objects.RegisterObject("webhooks", webhooksObject{})
	printer_objects.RegisterObject("configfile", configFileObject{})

	go system_info.Run()
	go temp_store.Run()
	go Connect()
}

func SetState(newState KlippyState, message string) {
	stateMutex.Lock()
	state, stateMessage = newState, message
	stateMutex.Unlock()

	switch newState {
	case Error:
		log.Errorln(message)
	default:
		log.Println(message)
	}

	if newState == Ready || newState == Shutdown {
		notify := notification.New("notify_klippy_"+string(newState), []any{})
		if err := notification.Publish(notify); err != nil {
			log.Errorf("Failed to public notification: %v", err)
		}
	}

	if err := printer_objects.EmitObject("webhooks"); err != nil {
		log.Errorf("Failed to emit object webhooks: %v", err)
	}
}

// Connect connects to the printer unless a connection attempt is already
// running or a printer is already connected. It returns once the printer
// disconnects.
func Connect() {
	if !connectMutex.TryLock() {
		return
	}
	defer connectMutex.Unlock()
	connect()
}

// Reconnect disconnects the current printer, if any, and connects again in
// the background.
func Reconnect() error {
	if p := GetPrinter(); p != nil {
		if err := p.Disconnect(); err != nil {
			return err
		}
	}
	go func() {
		// Wait for the previous connection to be torn down.
		connectMutex.Lock()
		defer connectMutex.Unlock()
		connect()
	}()
	return nil
}

func connect() {

	SetState(Startup, "Connecting to printer...")

	var baudRateInt int
	port := Config.Serial.Port
	// TOML integers decode as int64
	switch baudRate := Config.Serial.BaudRate.(type) {
	case int64:
		baudRateInt = int(baudRate)
	case int:
		baudRateInt = baudRate
	}
	if port == "" || port == "auto" || baudRateInt <= 0 {
		port, baudRateInt = scanner.FindSerialPort(Config)
	}

	if port == "" || baudRateInt == 0 {
		SetState(Error, "Could not find serial port to connect to")
		return
	}

	log.Printf("Using port %s @ %d", port, baudRateInt)

	p, err := printer.New(Config, port, baudRateInt)
	if err != nil {
		SetState(Error, "Error: "+err.Error())
		return
	}

	currentPrinter.Store(p)
	SetState(Ready, "Printer is ready")
	<-p.CloseCh
	currentPrinter.Store(nil)
	temp_store.Reset()
	if err := p.GetError(); err != nil {
		SetState(Error, err.Error())
	} else {
		SetState(Shutdown, "Disconnected from printer")
	}
}
