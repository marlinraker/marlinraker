package temp_store

import (
	"marlinraker-go/src/printer/parser"
	"sync"
	"time"
)

type tempRecords struct {
	Temperatures []float32 `json:"temperatures,omitempty"`
	Targets      []float32 `json:"targets,omitempty"`
	Powers       []float32 `json:"powers,omitempty"`
}

type TempStore map[string]tempRecords

var (
	store             = make(TempStore)
	storeMutex        = &sync.RWMutex{}
	lastMeasured      map[string]any
	lastMeasuredMutex = &sync.RWMutex{}
)

func Run() {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		storeTemps()
	}
}

func Reset() {
	storeMutex.Lock()
	store = make(map[string]tempRecords)
	storeMutex.Unlock()
}

func GetStore() TempStore {
	storeMutex.RLock()
	defer storeMutex.RUnlock()
	return store
}

func SetLastMeasured(_lastMeasured map[string]any) {
	lastMeasuredMutex.Lock()
	defer lastMeasuredMutex.Unlock()
	lastMeasured = _lastMeasured
}

func storeTemps() {
	storeMutex.Lock()
	lastMeasuredMutex.RLock()
	defer storeMutex.Unlock()
	defer lastMeasuredMutex.RUnlock()

	for name, temp := range lastMeasured {

		records, exist := store[name]
		if !exist {
			records = tempRecords{}
			store[name] = records
		}

		switch temp := temp.(type) {

		case parser.Sensor:
			appendToRecord(&records.Temperatures, float32(temp.Temperature))

		case parser.Heater:
			appendToRecord(&records.Temperatures, float32(temp.Temperature))
			appendToRecord(&records.Targets, float32(temp.Target))
			appendToRecord(&records.Powers, float32(temp.Power))
		}

		store[name] = records
	}
}

func appendToRecord(record *[]float32, value float32) {
	if *record == nil {
		*record = make([]float32, 1200)
	}
	*record = append(*record, value)[1:]
}
