package executors

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/printer_objects"
	"marlinraker/src/system_info/procfs"
	"marlinraker/src/util"
	"net/http"
	"strings"
)

type PrinterObjectsSubscribeResult struct {
	EventTime float64                                `json:"eventtime"`
	Status    map[string]printer_objects.QueryResult `json:"status"`
}

func PrinterObjectsSubscribeHttp(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	connectionId, err := params.RequireInt64("connection_id")
	if err != nil {
		return nil, err
	}

	connection, found := connections.FindConnection(int(connectionId))

	if !found {
		return nil, util.NewErrorf(400, "connection with id %d does not exist", connectionId)
	}

	subscriptions := make(map[string][]string)
	for name, attributes := range params {
		attributes := fmt.Sprintf("%v", attributes)
		if name == "connection_id" {
			continue
		}
		if attributes == "" {
			subscriptions[name] = nil
		} else {
			subscriptions[name] = strings.Split(attributes, ",")
		}
	}

	return subscribe(connection, subscriptions)
}

func PrinterObjectsSubscribeSocket(connection *connections.Connection, _ *http.Request, params Params) (any, error) {
	objects, exist := params["objects"].(map[string]any)
	if !exist {
		return nil, util.NewError(400, "objects param is required")
	}

	subscriptions := make(map[string][]string)
	for name, attributes := range objects {
		if attributes == nil {
			subscriptions[name] = nil
		} else if attributes, isArray := attributes.([]any); isArray {
			subscriptions[name] = lo.Map(attributes, func(item any, _ int) string {
				return fmt.Sprintf("%v", item)
			})
		}
	}
	return subscribe(connection, subscriptions)
}

func subscribe(connection *connections.Connection, subscriptions map[string][]string) (any, error) {

	eventTime, err := procfs.GetUptime()
	if err != nil {
		return nil, err
	}

	results := PrinterObjectsSubscribeResult{
		EventTime: eventTime,
		Status:    make(map[string]printer_objects.QueryResult),
	}

	if len(subscriptions) == 0 {
		printer_objects.Unsubscribe(connection)
		return results, nil
	}

	var errs []error
	for name, attributes := range subscriptions {
		printer_objects.Subscribe(connection, name, attributes)
		result, err := printer_objects.Query(name)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to query %s: %v", name, err))
			continue
		}

		if attributes != nil {
			filteredResult := make(printer_objects.QueryResult)
			for _, attribute := range attributes {
				if value, exists := result[attribute]; exists {
					filteredResult[attribute] = value
				}
			}
			result = filteredResult
		}
		results.Status[name] = result
	}

	if err = errors.Join(errs...); err != nil {
		return nil, err
	}
	return results, nil
}
