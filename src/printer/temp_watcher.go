package printer

import (
	"marlinraker/src/marlinraker/temp_store"
	"marlinraker/src/printer/parser"
	"marlinraker/src/printer_objects"
	"marlinraker/src/util"
	"math"
	"regexp"
	"sync"
	"time"
)

type heaterObject struct {
	mutex       *sync.RWMutex
	temperature float64
	target      float64
	power       float64
}

func (heaterObject *heaterObject) Query() printer_objects.QueryResult {
	heaterObject.mutex.RLock()
	defer heaterObject.mutex.RUnlock()
	return printer_objects.QueryResult{
		"temperature": heaterObject.temperature,
		"target":      heaterObject.target,
		"power":       heaterObject.power,
	}
}

type heatersObject struct {
	availableHeaters []string
	availableSensors []string
}

func (heatersObject heatersObject) Query() printer_objects.QueryResult {
	return printer_objects.QueryResult{
		"available_heaters": heatersObject.availableHeaters,
		"available_sensors": heatersObject.availableSensors,
	}
}

type temperatureSensorObject struct {
	mutex           *sync.RWMutex
	temperature     float64
	measuredMinTemp float64
	measuredMaxTemp float64
}

func (temperatureSensorObject *temperatureSensorObject) Query() printer_objects.QueryResult {
	temperatureSensorObject.mutex.RLock()
	defer temperatureSensorObject.mutex.RUnlock()
	return printer_objects.QueryResult{
		"temperature":       temperatureSensorObject.temperature,
		"measured_min_temp": temperatureSensorObject.measuredMinTemp,
		"measured_max_temp": temperatureSensorObject.measuredMaxTemp,
	}
}

type tempWatcher struct {
	printer       *Printer
	heatersCh     chan heatersObject
	firstRead     bool
	ticker        *time.Ticker
	heaterObjects map[string]*heaterObject
	sensorObjects map[string]*temperatureSensorObject
	objectsMutex  *sync.Mutex
	autoReport    bool
}

var (
	tempResponseLineRegex = regexp.MustCompile(`\s*T:`)
)

func newTempWatcher(printer *Printer) *tempWatcher {

	watcher := &tempWatcher{
		printer:       printer,
		heatersCh:     make(chan heatersObject),
		heaterObjects: make(map[string]*heaterObject),
		sensorObjects: make(map[string]*temperatureSensorObject),
		objectsMutex:  &sync.Mutex{},
		autoReport:    printer.Capabilities["AUTOREPORT_TEMP"],
		firstRead:     true,
	}

	if watcher.autoReport && !printer.IsPrusa {
		_ = printer.context.QueueGcode("M115 S1", true)
	}

	go watcher.runTimer()
	return watcher
}

func (watcher *tempWatcher) runTimer() {
	watcher.ticker = time.NewTicker(time.Second)
	for {
		<-watcher.ticker.C
		watcher.tick()
	}
}

func (watcher *tempWatcher) tick() {
	watcher.objectsMutex.Lock()
	defer watcher.objectsMutex.Unlock()

	for name := range watcher.heaterObjects {
		err := printer_objects.EmitObject(name)
		if err != nil {
			util.LogError(err)
		}
	}

	for name := range watcher.sensorObjects {
		err := printer_objects.EmitObject(name)
		if err != nil {
			util.LogError(err)
		}
	}

	if !watcher.autoReport {
		watcher.parseTemps(<-watcher.printer.context.QueueGcode("M105", true))
	}
}

func (watcher *tempWatcher) stop() {
	if watcher.ticker != nil {
		watcher.ticker.Stop()
	}
	printer_objects.UnregisterObject("heaters")
	for name := range watcher.heaterObjects {
		printer_objects.UnregisterObject(name)
	}
	for name := range watcher.sensorObjects {
		printer_objects.UnregisterObject(name)
	}
}

func (watcher *tempWatcher) handle(line string) bool {
	if tempResponseLineRegex.MatchString(line) {
		watcher.objectsMutex.Lock()
		defer watcher.objectsMutex.Unlock()
		watcher.parseTemps(line)
		return true
	}
	return false
}

func (watcher *tempWatcher) parseTemps(data string) {

	temps, err := parser.ParseM105(data)
	if err != nil {
		util.LogError(err)
		return
	}

	if watcher.firstRead {
		watcher.heatersCh <- watcher.setupHeaters(temps)
		watcher.firstRead = false
	}

	temp_store.SetLastMeasured(temps)

	for name, temp := range temps {

		switch temp := temp.(type) {
		case parser.Heater:
			if heater, exists := watcher.heaterObjects[name]; exists {
				heater.mutex.Lock()
				heater.temperature, heater.target, heater.power =
					temp.Temperature, temp.Target, temp.Power
				heater.mutex.Unlock()
			}

		case parser.Sensor:
			if sensor, exists := watcher.sensorObjects[name]; exists {
				sensor.mutex.Lock()
				sensor.temperature, sensor.measuredMinTemp, sensor.measuredMaxTemp =
					temp.Temperature,
					math.Min(temp.Temperature, sensor.measuredMinTemp),
					math.Max(temp.Temperature, sensor.measuredMaxTemp)
				sensor.mutex.Unlock()
			}
		}
	}
}

func (watcher *tempWatcher) setupHeaters(temps map[string]any) heatersObject {

	heaters, sensors := make([]string, 0), make([]string, 0)

	for name, temp := range temps {
		sensors = append(sensors, name)
		switch temp := temp.(type) {

		case parser.Heater:
			heaters = append(heaters, name)
			obj := &heaterObject{&sync.RWMutex{}, temp.Temperature, temp.Target, temp.Power}
			printer_objects.RegisterObject(name, obj)
			watcher.heaterObjects[name] = obj

		case parser.Sensor:
			obj := &temperatureSensorObject{&sync.RWMutex{}, temp.Temperature, temp.Temperature, temp.Temperature}
			printer_objects.RegisterObject(name, obj)
			watcher.sensorObjects[name] = obj
		}
	}

	obj := heatersObject{heaters, sensors}
	printer_objects.RegisterObject("heaters", obj)
	return obj
}
