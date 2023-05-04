package executors

import (
	"fmt"
	"marlinraker-go/src/marlinraker/connections"
	"marlinraker-go/src/printer_objects"
	"net/http"
	"strings"
	"time"
)

type PrinterObjectsQueryResult struct {
	EventTime float64                                `json:"eventtime"`
	Status    map[string]printer_objects.QueryResult `json:"status"`
}

func PrinterObjectsQuery(_ *connections.Connection, _ *http.Request, params Params) (any, error) {

	results := PrinterObjectsQueryResult{
		EventTime: float64(time.Now().UnixMilli()) / 1000.0,
		Status:    make(map[string]printer_objects.QueryResult),
	}

	for name, attributes := range params {
		attributes := fmt.Sprintf("%v", attributes)
		result, err := printer_objects.Query(name)
		if err != nil {
			return nil, err
		}
		if attributes != "" {
			filteredResult := make(printer_objects.QueryResult)
			for _, attribute := range strings.Split(attributes, ",") {
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
