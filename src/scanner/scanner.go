package scanner

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
	"marlinraker/src/config"
	"marlinraker/src/database"
	"marlinraker/src/util"
	"strconv"
	"time"
)

func FindSerialPort(config *config.Config) (string, int) {

	lastPath, _ := database.GetItem("marlinraker", "lastPath", true)
	lastBaudRate, _ := database.GetItem("marlinraker", "lastBaudRate", true)

	path, hasPath := lastPath.(string)
	baudRateFloat, hasBaudRate := lastBaudRate.(float64)
	if hasPath && hasBaudRate {

		baudRate := int(baudRateFloat)
		log.Println("Trying last used port " + path + " @ " + strconv.Itoa(baudRate) + "...")

		success := tryPort(path, baudRate, config.Serial.ConnectionTimeout)
		if success {
			log.Println("Found printer at last used port " + path + " @ " + strconv.Itoa(baudRate))
			return path, baudRate
		}
		time.Sleep(time.Millisecond * 500)
	}

	path, baudRate := scan(config)
	if path != "" && baudRate != 0 {
		_, err := database.PostItem("marlinraker", "lastPath", path, true)
		if err != nil {
			util.LogError(err)
		}
		_, err = database.PostItem("marlinraker", "lastBaudRate", baudRate, true)
		if err != nil {
			util.LogError(err)
		}
		log.Debugln("Saved last successful port " + path + " @ " + strconv.Itoa(baudRate))
	}
	return path, baudRate
}

func scan(config *config.Config) (string, int) {

	var err error
	var ports []string
	if config.Serial.Port == "" || config.Serial.Port == "auto" {
		ports, err = serial.GetPortsList()
		if err != nil {
			util.LogError(err)
			return "", 0
		}
	} else {
		ports = []string{config.Serial.Port}
	}

	var baudRates []int
	switch baudRate := config.Serial.BaudRate.(type) {
	case int:
		baudRates = []int{baudRate}
		break
	default:
		baudRates = []int{250000, 115200, 19200}
		break
	}

	for _, path := range ports {
		for _, baudRate := range baudRates {
			log.Println("Trying port " + path + " @ " + strconv.Itoa(baudRate) + "...")
			success := tryPort(path, baudRate, config.Serial.ConnectionTimeout)
			if success {
				log.Println("Found printer at " + path + " @ " + strconv.Itoa(baudRate))
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
		log.Error(err)
		return false
	}

	defer func(port serial.Port) {
		err := port.Close()
		if err != nil {
			util.LogError(err)
		}
	}(port)

	if _, err = port.Write([]byte("M110 N0\n")); err != nil {
		log.Error(err)
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
				log.WithField("port", path).WithField("baudRate", baudRate).Debugln("recv: " + line)
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
		log.Errorln("Timeout on " + path)
		if err := port.Close(); err != nil {
			util.LogError(err)
		}
	case success = <-connectCh:
	}
	return success
}
