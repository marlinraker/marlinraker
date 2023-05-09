package macros

import (
	"errors"
	"marlinraker/src/printer_objects"
	"marlinraker/src/shared"
	"strconv"
	"strings"
)

type MacroManager struct {
	printer shared.Printer
	macros  map[string]Macro
}

type Params map[string]string

func (params Params) RequireString(name string) (string, error) {
	value, exists := params[name]
	if !exists {
		return "", errors.New("missing argument " + strings.ToUpper(name))
	}
	return value, nil
}

func (params Params) RequireFloat64(name string) (float64, error) {
	value, err := params.RequireString(name)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(value, 64)
}

type Objects map[string]printer_objects.QueryResult

type Macro interface {
	Description() string
	Execute(*MacroManager, []string, Objects, Params) error
}

func NewMacroManager(printer shared.Printer) *MacroManager {
	return &MacroManager{
		printer: printer,
		macros: map[string]Macro{
			"SET_HEATER_TEMPERATURE": setHeaterTemperatureMacro{},
		},
	}
}

func (manager *MacroManager) TryCommand(command string) chan error {

	parts := strings.Split(command, " ")
	name := strings.ToUpper(parts[0])

	if macro, exists := manager.macros[name]; exists {
		objects, params := make(Objects), make(Params)
		for name, object := range printer_objects.GetObjects() {
			objects[name] = object.Query()
		}

		rawParams := parts[1:]
		for _, rawParam := range rawParams {
			if idx := strings.Index(rawParam, "="); idx != -1 {
				name, value := rawParam[:idx], rawParam[idx+1:]
				if name != "" && value != "" {
					params[strings.ToLower(name)] = value
				}
			}
		}

		ch := make(chan error)
		go func() {
			defer close(ch)
			ch <- macro.Execute(manager, rawParams, objects, params)
		}()
		return ch
	}

	return nil
}
