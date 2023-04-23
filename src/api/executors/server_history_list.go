// TODO

package executors

import "marlinraker-go/src/marlinraker/connections"

type ServerHistoryListResult struct {
	Count int   `json:"count"`
	Jobs  []any `json:"jobs"`
}

func ServerHistoryList(*connections.Connection, Params) (any, error) {
	return ServerHistoryListResult{
		Count: 0, Jobs: make([]any, 0),
	}, nil
}
