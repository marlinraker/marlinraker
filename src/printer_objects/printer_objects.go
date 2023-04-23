package printer_objects

import (
	"errors"
	"marlinraker-go/src/api/notification"
	"marlinraker-go/src/marlinraker/connections"
	"reflect"
	"sync"
	"time"
)

type QueryResult map[string]any

type PrinterObject interface {
	Query() QueryResult
}

type Subscriptions map[*connections.Connection][]string

var (
	objects            = make(map[string]PrinterObject)
	objectsMutex       = &sync.RWMutex{}
	subscriptions      = make(map[string]Subscriptions)
	subscriptionsMutex = &sync.RWMutex{}
	lastEmitted        = make(map[*connections.Connection]map[string]QueryResult)
)

func GetObjects() map[string]PrinterObject {
	objectsMutex.RLock()
	defer objectsMutex.RUnlock()
	return objects
}

func Query(name string) (QueryResult, error) {
	objectsMutex.RLock()
	defer objectsMutex.RUnlock()

	object, exists := objects[name]
	if !exists {
		return nil, errors.New("object with name \"" + name + "\" does not exist")
	}
	return object.Query(), nil
}

func EmitObject(name string) error {

	subscriptionsMutex.RLock()
	defer subscriptionsMutex.RUnlock()

	result, err := Query(name)
	if err != nil {
		return err
	}
	eventTime := float64(time.Now().UnixMilli()) / 1000.0

	for connection, attributes := range subscriptions[name] {
		diff := getDiff(connection, name, result)
		if attributes != nil {
			filtered := make(QueryResult)
			for _, attribute := range attributes {
				if value, exists := diff[attribute]; exists {
					filtered[attribute] = value
				}
			}
			diff = filtered
		}
		status := map[string]any{name: diff}
		err := notification.Send(connection, notification.New("notify_status_update", []any{status, eventTime}))
		if err != nil {
			return err
		}
	}
	return nil
}

func RegisterObject(name string, object PrinterObject) {
	objectsMutex.Lock()
	defer objectsMutex.Unlock()
	objects[name] = object
}

func UnregisterObject(name string) {
	objectsMutex.Lock()
	defer objectsMutex.Unlock()
	delete(objects, name)
}

func Subscribe(connection *connections.Connection, name string, attributes []string) {
	subscriptionsMutex.Lock()
	defer subscriptionsMutex.Unlock()

	if _, exists := subscriptions[name]; !exists {
		subscriptions[name] = make(Subscriptions)
	}
	subscriptions[name][connection] = attributes
}

func Unsubscribe(connection *connections.Connection) {
	subscriptionsMutex.Lock()
	defer subscriptionsMutex.Unlock()

	for name, subscription := range subscriptions {
		for _connection := range subscription {
			if _connection == connection {
				delete(subscription, _connection)
			}
		}
		subscriptions[name] = subscription
	}
}

func getDiff(connection *connections.Connection, name string, result QueryResult) QueryResult {

	last, exists := lastEmitted[connection][name]
	if _, exists := lastEmitted[connection]; !exists {
		lastEmitted[connection] = make(map[string]QueryResult)
	}
	lastEmitted[connection][name] = result

	if !exists {
		return result
	}

	diff := make(map[string]any)
	for key, value := range result {
		if lastValue, exists := last[key]; !exists || !reflect.DeepEqual(value, lastValue) {
			diff[key] = value
		}
	}

	return diff
}
