package macros

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/config"
	"marlinraker/src/printer_objects"
	"marlinraker/src/shared"
	"strconv"
	"strings"
)

type MacroManager struct {
	Macros  map[string]Macro
	printer shared.Printer
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

func NewMacroManager(printer shared.Printer, config *config.Config) *MacroManager {

	macros := map[string]Macro{
		"CANCEL_PRINT":           cancelPrintMacro{},
		"PAUSE":                  pauseMacro{},
		"RESUME":                 resumeMacro{},
		"SDCARD_PRINT_FILE":      sdcardPrintFileMacro{},
		"SDCARD_RESET_FILE":      sdcardResetFileMacro{},
		"SET_HEATER_TEMPERATURE": setHeaterTemperatureMacro{},
		"TURN_OFF_HEATERS":       turnOffHeatersMacro{},
	}

	for name, macroConfig := range config.Macros {
		name = strings.ToUpper(name)
		macro, err := newCustomMacro(name, "G-Code macro", macroConfig.Gcode)
		if err != nil {
			log.Errorln("Error while loading macro \"" + name + "\": " + err.Error())
			continue
		}
		if existing, exists := macros[name]; exists {
			rename := strings.ToUpper(macroConfig.RenameExisting)
			if rename == "" {
				rename = name + "_BASE"
			}
			if _, exists := macros[rename]; exists {
				log.Errorln("Error while loading macro \"" + name + "\": Macro \"" + rename + "\" already exists." +
					" Choose another macro name with \"rename_existing\"")
				continue
			}
			macros[rename] = renamedMacro{
				original:    existing,
				description: "Renamed builtin of '" + name + "'",
			}
		} else if macroConfig.RenameExisting != "" {
			log.Warningln("Warning while loading macro \"" + name + "\": \"rename_existing\" was specified " +
				"although a macro with the name \"" + name + "\" did not exist before")
		}
		macros[name] = macro
	}

	return &MacroManager{macros, printer}
}

func (manager *MacroManager) TryCommand(command string) chan error {

	parts := strings.Split(command, " ")
	name := strings.ToUpper(parts[0])

	if macro, exists := manager.Macros[name]; exists {
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
