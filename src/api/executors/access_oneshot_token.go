package executors

import (
	"marlinraker/src/auth"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type AccessOneshotTokenResult string

func AccessOneshotToken(*connections.Connection, *http.Request, Params) (any, error) {
	return auth.GenerateOneshotToken()
}
