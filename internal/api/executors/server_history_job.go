package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/history"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/util"
)

type ServerHistoryGetJobResult struct {
	Job history.Job `json:"job"`
}

func ServerHistoryGetJob(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	uid, err := params.RequireString("uid")
	if err != nil {
		return nil, err
	}
	job, err := history.GetJob(uid)
	if err != nil {
		return nil, err
	}
	return ServerHistoryGetJobResult{Job: job}, nil
}

type ServerHistoryDeleteJobResult struct {
	DeletedJobs []string `json:"deleted_jobs"`
}

func ServerHistoryDeleteJob(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	all, _ := params.GetBool("all")
	uid, _ := params.GetString("uid")
	if !all && uid == "" {
		return nil, util.NewError(400, "uid or all param is required")
	}
	deleted, err := history.DeleteJob(uid, all)
	if err != nil {
		return nil, err
	}
	return ServerHistoryDeleteJobResult{DeletedJobs: deleted}, nil
}

type ServerHistoryTotalsResult struct {
	JobTotals       history.Totals `json:"job_totals"`
	AuxiliaryTotals []any          `json:"auxiliary_totals"`
}

func ServerHistoryTotals(*connections.Connection, *http.Request, Params) (any, error) {
	return ServerHistoryTotalsResult{JobTotals: history.GetTotals(), AuxiliaryTotals: []any{}}, nil
}

type ServerHistoryResetTotalsResult struct {
	LastTotals          history.Totals `json:"last_totals"`
	LastAuxiliaryTotals []any          `json:"last_auxiliary_totals"`
}

func ServerHistoryResetTotals(*connections.Connection, *http.Request, Params) (any, error) {
	last, err := history.ResetTotals()
	if err != nil {
		return nil, err
	}
	return ServerHistoryResetTotalsResult{LastTotals: last, LastAuxiliaryTotals: []any{}}, nil
}
