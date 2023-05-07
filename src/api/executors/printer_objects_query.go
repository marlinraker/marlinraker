package executors

import (
	"fmt"
	"github.com/samber/lo"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/printer_objects"
	"marlinraker/src/util"
	"net/http"
	"strings"
	"time"
)

type PrinterObjectsQueryResult struct {
	EventTime float64                                `json:"eventtime"`
	Status    map[string]printer_objects.QueryResult `json:"status"`
}

func PrinterObjectsQueryHttp(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	results := PrinterObjectsQueryResult{
		EventTime: float64(time.Now().UnixMilli()) / 1000.0,
		Status:    make(map[string]printer_objects.QueryResult),
	}

	for name, attributes := range params {
		attributesStr := fmt.Sprintf("%v", attributes)
		var attributes []string = nil
		if attributesStr != "" {
			attributes = strings.Split(attributesStr, ",")
		}
		if result, err := query(name, attributes); err == nil {
			results.Status[name] = result
		} else {
			return nil, err
		}
	}

	return results, nil
}

func PrinterObjectsQuerySocket(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	objects, exist := params["objects"].(map[string]any)
	if !exist {
		return nil, util.NewError("objects param is required", 400)
	}

	results := PrinterObjectsQueryResult{
		EventTime: float64(time.Now().UnixMilli()) / 1000.0,
		Status:    make(map[string]printer_objects.QueryResult),
	}

	for name, attributes := range objects {
		var attributesStr []string = nil
		if attributes != nil {
			if attributes, isArray := attributes.([]any); isArray {
				attributesStr = lo.Map(attributes, func(attr any, _ int) string {
					return fmt.Sprintf("%v", attr)
				})
			} else {
				return nil, util.NewError("subscribed topics have to be nil or string list", 400)
			}
		}
		if result, err := query(name, attributesStr); err == nil {
			results.Status[name] = result
		} else {
			return nil, err
		}
	}

	return results, nil
}

func query(name string, attributes []string) (printer_objects.QueryResult, error) {
	result, err := printer_objects.Query(name)
	if err != nil {
		return nil, err
	}
	if len(attributes) > 0 {
		filteredResult := make(printer_objects.QueryResult)
		for _, attribute := range attributes {
			if value, exists := result[attribute]; exists {
				filteredResult[attribute] = value
			}
		}
		result = filteredResult
	}
	return result, nil
}
