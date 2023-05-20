package marlinraker

import (
	log "github.com/sirupsen/logrus"
	"marlinraker/src/api/notification"
	"marlinraker/src/config"
	"marlinraker/src/constants"
	"marlinraker/src/marlinraker/temp_store"
	"marlinraker/src/printer"
	"marlinraker/src/printer_objects"
	"marlinraker/src/scanner"
	"marlinraker/src/system_info"
	"marlinraker/src/util"
	"path/filepath"
	"strconv"
)

type KlippyState string

const (
	Ready    KlippyState = "ready"
	Error    KlippyState = "error"
	Shutdown KlippyState = "shutdown"
	Startup  KlippyState = "startup"
)

var (
	State           = Shutdown
	StateMessage    string
	Config          *config.Config
	KlipperSettings map[string]any
	KlipperConfig   map[string]any
	Printer         *printer.Printer
)

func Init(dataDir string) {
	log.Println("Starting Marlinraker " + constants.Version)

	configDir := filepath.Join(dataDir, "config")

	var err error
	if err = config.CopyDefaults(configDir); err != nil {
		panic(err)
	}

	Config, err = config.LoadConfig(filepath.Join(configDir, "marlinraker.toml"))
	if err != nil {
		panic(err)
	}
	KlipperSettings, KlipperConfig = config.GenerateFakeKlipperConfig(Config)

	if Config.Misc.ExtendedLogs {
		log.SetLevel(log.DebugLevel)
	}

	printer_objects.RegisterObject("webhooks", webhooksObject{})
	printer_objects.RegisterObject("configfile", configFileObject{})

	go system_info.Run()
	go temp_store.Run()
	go Connect()
}

func SetState(state KlippyState, message string) {
	State = state
	StateMessage = message

	switch state {
	case Error:
		log.Errorln(message)
	default:
		log.Println(message)
	}

	if state == Ready || state == Shutdown {
		notify := notification.New("notify_klippy_"+string(state), []any{})
		if err := notification.Publish(notify); err != nil {
			util.LogError(err)
		}
	}

	if err := printer_objects.EmitObject("webhooks"); err != nil {
		util.LogError(err)
	}
}

func Connect() {

	if State != Error && State != Shutdown {
		return
	}

	SetState(Startup, "Connecting to printer...")

	var baudRateInt int
	port, baudRate := Config.Serial.Port, Config.Serial.BaudRate
	if baudRate, isInt := baudRate.(int); isInt {
		baudRateInt = baudRate
	}
	if port == "" || port == "auto" || baudRateInt <= 0 {
		port, baudRateInt = scanner.FindSerialPort(Config)
	}

	if port == "" || baudRateInt == 0 {
		SetState(Error, "Could not find serial port to connect to")
		return
	}

	log.Println("Using port " + port + " @ " + strconv.Itoa(baudRateInt))

	var err error
	Printer, err = printer.New(Config, port, baudRateInt)
	if err != nil {
		SetState(Error, "Error: "+err.Error())
		Printer = nil
		return
	}

	SetState(Ready, "Printer is ready")
	<-Printer.CloseCh
	temp_store.Reset()
	if Printer.Error != nil {
		SetState(Error, Printer.Error.Error())
	} else {
		SetState(Shutdown, "Disconnected from printer")
	}
	Printer = nil
}
