package service

import (
	"context"
	"errors"
	"github.com/coreos/go-systemd/v22/dbus"
	"marlinraker/src/config"
	"strings"
)

type UnitState struct {
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
}

type Action int

const (
	Start Action = iota
	Restart
	Stop
)

var (
	conn            *dbus.Conn
	allowedServices []string
	errNoConn       = errors.New("no dbus connection")
	ctx             = context.Background()
)

func Init(config *config.Config) error {
	var err error
	conn, err = dbus.NewWithContext(ctx)

	allowedServices = config.Misc.AllowedServices
	for i, service := range allowedServices {
		if !strings.HasSuffix(service, ".service") {
			allowedServices[i] = service + ".service"
		}
	}
	return err
}

func Close() {
	if conn != nil {
		conn.Close()
	}
}

func GetServiceState() (map[string]UnitState, error) {
	if conn == nil {
		return nil, errNoConn
	}

	units, err := conn.ListUnitsByNamesContext(ctx, allowedServices)
	if err != nil {
		return nil, err
	}

	state := make(map[string]UnitState)
	for _, unit := range units {
		if unit.LoadState == "not-found" {
			continue
		}
		state[strings.TrimSuffix(unit.Name, ".service")] = UnitState{
			ActiveState: unit.ActiveState,
			SubState:    unit.SubState,
		}
	}
	return state, nil
}

func PerformAction(svc string, action Action) error {
	if conn == nil {
		return errNoConn
	}

	if !strings.HasSuffix(svc, ".service") {
		svc += ".service"
	}

	if svc == "marlinraker.service" && action != Restart {
		return errors.New("action not allowed on Marlinraker service")
	}

	ch := make(chan string)
	var f func(context.Context, string, string, chan<- string) (int, error)

	switch action {
	case Start:
		f = conn.StartUnitContext
	case Restart:
		f = conn.RestartUnitContext
	case Stop:
		f = conn.StopUnitContext
	}

	if _, err := f(ctx, svc, "fail", ch); err != nil {
		return err
	}
	<-ch
	return nil
}
