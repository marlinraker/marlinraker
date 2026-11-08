package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/history"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerHistoryListResult struct {
	Count int           `json:"count"`
	Jobs  []history.Job `json:"jobs"`
}

func ServerHistoryList(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	options := history.ListOptions{Limit: 50}
	if start, exists := params.GetInt64("start"); exists {
		options.Start = int(start)
	}
	if limit, exists := params.GetInt64("limit"); exists {
		options.Limit = int(limit)
	}
	if since, exists := params.GetFloat64("since"); exists {
		options.Since = since
	}
	if before, exists := params.GetFloat64("before"); exists {
		options.Before = before
	}
	if order, exists := params.GetString("order"); exists {
		options.Asc = order == "asc"
	}

	jobs, count := history.List(options)
	return ServerHistoryListResult{Count: count, Jobs: jobs}, nil
}
