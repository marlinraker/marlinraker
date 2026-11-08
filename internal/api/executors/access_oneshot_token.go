package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/auth"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type AccessOneshotTokenResult string

func AccessOneshotToken(*connections.Connection, *http.Request, Params) (any, error) {
	return auth.GenerateOneshotToken()
}
