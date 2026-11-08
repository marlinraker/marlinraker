package printer_objects

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/marlinraker/marlinraker/internal/api/notification"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/system_info/procfs"
)

type QueryResult map[string]any

type PrinterObject interface {
	Query() (QueryResult, error)
}

type Subscriptions map[*connections.Connection][]string

var (
	objects            = make(map[string]PrinterObject)
	objectsMutex       = &sync.RWMutex{}
	subscriptions      = make(map[string]Subscriptions)
	subscriptionsMutex = &sync.RWMutex{}
	lastEmitted        = make(map[*connections.Connection]map[string]QueryResult)
	lastEmittedMutex   = &sync.RWMutex{}
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
		return QueryResult{}, nil
	}
	return object.Query()
}

func EmitObject(names ...string) error {

	subscriptionsMutex.RLock()
	defer subscriptionsMutex.RUnlock()

	eventTime, err := procfs.GetUptime()
	if err != nil {
		return fmt.Errorf("failed to get uptime: %w", err)
	}

	pending := make(map[*connections.Connection]map[string]QueryResult)

	var (
		result QueryResult
		errs   []error
	)
	for _, name := range names {

		result, err = Query(name)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to query object: %w", err))
			continue
		}

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

			if _, exists := pending[connection]; !exists {
				pending[connection] = make(map[string]QueryResult)
			}
			pending[connection][name] = diff
		}
	}

	for connection, status := range pending {
		err = notification.Send(connection, notification.New("notify_status_update", []any{status, eventTime}))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to send notification: %w", err))
		}
	}
	return errors.Join(errs...)
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

	lastEmittedMutex.Lock()
	defer lastEmittedMutex.Unlock()
	delete(lastEmitted, connection)
}

func getDiff(connection *connections.Connection, name string, result QueryResult) QueryResult {

	lastEmittedMutex.Lock()
	last, exists := lastEmitted[connection][name]
	if _, exists := lastEmitted[connection]; !exists {
		lastEmitted[connection] = make(map[string]QueryResult)
	}
	lastEmitted[connection][name] = result
	lastEmittedMutex.Unlock()

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
