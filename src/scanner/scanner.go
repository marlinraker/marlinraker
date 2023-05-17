package scanner

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
	"marlinraker/src/config"
	"marlinraker/src/util"
	"strconv"
	"time"
)

func FindSerialPort(config *config.Config) (string, int64) {

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
				return path, int64(baudRate)
			}
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
	successCh1, successCh2 := make(chan bool), make(chan bool)

	go func() {
		defer close(successCh1)
		time.Sleep(time.Millisecond * time.Duration(connectionTimeout))
		successCh1 <- false
	}()

	go func() {
		defer close(successCh2)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				log.WithField("port", path).WithField("baudRate", baudRate).Debugln("recv: " + line)
			}
			if line == "ok" {
				successCh2 <- true
				return
			}
		}
		successCh2 <- false
	}()

	var success bool
	select {
	case success = <-successCh1:
	case success = <-successCh2:
	}
	return success
}
