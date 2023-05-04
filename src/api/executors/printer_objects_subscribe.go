package executors

import (
	"fmt"
	"github.com/samber/lo"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/printer_objects"
	"marlinraker-go/src/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PrinterObjectsSubscribeResult struct {
	EventTime float64                                `json:"eventtime"`
	Status    map[string]printer_objects.QueryResult `json:"status"`
}

func PrinterObjectsSubscribeHttp(connection *connections.Connection, _ *http.Request, params Params) (any, error) {

	connectionId, exists := params.GetInt64("connection_id")
	if !exists {
		return nil, util.NewError("connection_id param is required", 400)
	}

	connection, found := lo.Find(connections.GetConnections(), func(connection *connections.Connection) bool {
		return connection.Id == int(connectionId)
	})

	if !found {
		return nil, util.NewError("connection with id "+strconv.Itoa(int(connectionId))+" does not exist", 400)
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
		return nil, util.NewError("objects param is required", 400)
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

	results := PrinterObjectsSubscribeResult{
		EventTime: float64(time.Now().UnixMilli()) / 1000.0,
		Status:    make(map[string]printer_objects.QueryResult),
	}

	if len(subscriptions) == 0 {
		printer_objects.Unsubscribe(connection)
		return results, nil
	}

	for name, attributes := range subscriptions {
		printer_objects.Subscribe(connection, name, attributes)
		result, err := printer_objects.Query(name)
		if err != nil {
			return nil, err
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

	return results, nil
}
