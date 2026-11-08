package scanner

import (
	"bufio"
	"time"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"

	"github.com/marlinraker/marlinraker/internal/config"
	"github.com/marlinraker/marlinraker/internal/database"
)

var scanning = false

func FindSerialPort(config *config.Config) (string, int) {

	if scanning {
		return "", 0
	}
	scanning = true
	defer func() {
		scanning = false
	}()

	if config.Serial.Port == "auto" && config.Serial.BaudRate == "auto" {
		lastPath, _ := database.GetItem("marlinraker", "lastPath", true)
		lastBaudRate, _ := database.GetItem("marlinraker", "lastBaudRate", true)

		path, hasPath := lastPath.(string)
		baudRateFloat, hasBaudRate := lastBaudRate.(float64)
		if hasPath && hasBaudRate {

			baudRate := int(baudRateFloat)
			log.Printf("Trying last used port %s @ %d...", path, baudRate)

			success := tryPort(path, baudRate, config.Serial.ConnectionTimeout)
			if success {
				log.Printf("Found printer at last used port %s @ %d", path, baudRate)
				return path, baudRate
			}
			time.Sleep(time.Millisecond * 500)
		}
	}

	path, baudRate := scan(config)
	if path != "" && baudRate != 0 {
		_, err := database.PostItem("marlinraker", "lastPath", path, true)
		if err != nil {
			log.Errorf("Failed to post database item: %v", err)
		}
		_, err = database.PostItem("marlinraker", "lastBaudRate", baudRate, true)
		if err != nil {
			log.Errorf("Failed to post database item: %v", err)
		}
		log.Debugf("Saved last successful port %s @ %d", path, baudRate)
	}
	return path, baudRate
}

func scan(config *config.Config) (string, int) {

	var err error
	var ports []string
	if config.Serial.Port == "" || config.Serial.Port == "auto" {
		ports, err = serial.GetPortsList()
		if err != nil {
			log.Errorf("Failed to get serial port list: %v", err)
			return "", 0
		}
	} else {
		ports = []string{config.Serial.Port}
	}

	var baudRates []int
	switch baudRate := config.Serial.BaudRate.(type) {
	case int64:
		baudRates = []int{int(baudRate)}
	case int:
		baudRates = []int{baudRate}
	default:
		baudRates = []int{250000, 115200, 19200}
	}

	for _, path := range ports {
		for _, baudRate := range baudRates {
			log.Printf("Trying port %s @ %d...", path, baudRate)
			success := tryPort(path, baudRate, config.Serial.ConnectionTimeout)
			if success {
				log.Printf("Found printer at %s @ %d...", path, baudRate)
				return path, baudRate
			}
			// wait for serial port to recover
			time.Sleep(time.Millisecond * 500)
		}
	}

	return "", 0
}

func tryPort(path string, baudRate int, connectionTimeout int) bool {

	port, err := serial.Open(path, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		log.Errorf("Cannot open serial port: %v", err)
		return false
	}

	defer func(port serial.Port) {
		err := port.Close()
		if err != nil {
			log.Errorf("Failed to close serial port %q: %v", path, err)
		}
	}(port)

	if _, err = port.Write([]byte("M110 N0\n")); err != nil {
		log.Errorf("Failed to write to serial port: %v", err)
		return false
	}
	log.WithField("port", path).WithField("baudRate", baudRate).Debugln("write: M110 N0")

	scanner := bufio.NewScanner(port)
	timeoutCh, connectCh := make(chan bool), make(chan bool)

	go func() {
		defer close(timeoutCh)
		time.Sleep(time.Millisecond * time.Duration(connectionTimeout))
		timeoutCh <- false
	}()

	go func() {
		defer close(connectCh)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				log.WithField("port", path).
					WithField("baudRate", baudRate).
					Debugf("recv: %s", line)
			}
			if line == "ok" {
				connectCh <- true
				return
			}
		}
		connectCh <- false
	}()

	var success bool
	select {
	case success = <-timeoutCh:
		log.Errorf("Timeout on %s", path)
		if err := port.Close(); err != nil {
			log.Errorf("Failed to close port %q: %v", path, err)
		}
	case success = <-connectCh:
	}
	return success
}
