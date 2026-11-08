package macros

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/config"
	"github.com/marlinraker/marlinraker/internal/printer_objects"
	"github.com/marlinraker/marlinraker/internal/shared"
)

type MacroManager struct {
	Macros       map[string]Macro
	macroObjects []string
	printer      shared.Printer
}

type Params map[string]string

var (
	quotedParamRegex   = regexp.MustCompile(`(\S+)=("(?:[^"\\]|\\.)*?")`)
	unquotedParamRegex = regexp.MustCompile(`(\S+)=(\S+)`)
)

func (params Params) RequireString(name string) (string, error) {
	value, exists := params[name]
	if !exists {
		return "", fmt.Errorf("missing argument %s", strings.ToUpper(name))
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
	Execute(*MacroManager, shared.ExecutorContext, []string, Objects, Params) error
}

func NewMacroManager(printer shared.Printer, config *config.Config) *MacroManager {

	macros := map[string]Macro{
		"CANCEL_PRINT":           cancelPrintMacro{},
		"PAUSE":                  pauseMacro{},
		"RESTORE_GCODE_STATE":    restoreGcodeState{},
		"RESUME":                 resumeMacro{},
		"SAVE_GCODE_STATE":       saveGcodeState{},
		"SDCARD_PRINT_FILE":      sdcardPrintFileMacro{},
		"SDCARD_RESET_FILE":      sdcardResetFileMacro{},
		"SET_HEATER_TEMPERATURE": setHeaterTemperatureMacro{},
		"TURN_OFF_HEATERS":       turnOffHeatersMacro{},
	}

	var macroObjects []string

	for name, macroConfig := range config.Macros {
		name = strings.ToUpper(name)
		macro, err := newCustomMacro(name, "G-Code macro", macroConfig.Gcode)
		if err != nil {
			log.Errorf("Error while loading macro %q: %v", name, err)
			continue
		}
		if existing, exists := macros[name]; exists {
			rename := strings.ToUpper(macroConfig.RenameExisting)
			if rename == "" {
				rename = fmt.Sprintf("%s_BASE", name)
			}
			if _, exists := macros[rename]; exists {
				log.Errorf("Error while loading macro %q: Macro %q already exists."+
					" Choose another macro name with \"rename_existing\"", name, rename)
				continue
			}
			macros[rename] = renamedMacro{
				original:    existing,
				description: fmt.Sprintf("Original builtin of '%s'", name),
			}
		} else if macroConfig.RenameExisting != "" {
			log.Warningf("Warning while loading macro %q: \"rename_existing\" was specified "+
				"although a macro with the name %q did not exist before", name, name)
		}
		macros[name] = macro

		if macroConfig.Variables == nil {
			macroConfig.Variables = map[string]any{}
		}
		object, objectName := gcodeMacroObject{macroConfig.Variables}, "gcode_macro "+name
		printer_objects.RegisterObject(objectName, object)
		macroObjects = append(macroObjects, objectName)
	}

	return &MacroManager{macros, macroObjects, printer}
}

func (manager *MacroManager) Cleanup() {
	for _, objectName := range manager.macroObjects {
		printer_objects.UnregisterObject(objectName)
	}
}

func (manager *MacroManager) GetMacro(command string) (Macro, string, bool) {
	idx := strings.Index(command, " ")
	if idx != -1 {
		command = command[:idx]
	}
	command = strings.ToUpper(command)
	macro, exists := manager.Macros[command]
	return macro, command, exists
}

func (manager *MacroManager) ExecuteMacro(macro Macro, context shared.ExecutorContext, gcode string) (chan error, error) {

	var err error
	objects, params := make(Objects), make(Params)
	for name, object := range printer_objects.GetObjects() {
		objects[name], err = object.Query()
		if err != nil {
			return nil, err
		}
	}

	parts := strings.Split(gcode, " ")
	rawParams := parts[1:]

	for _, match := range quotedParamRegex.FindAllStringSubmatch(gcode, -1) {
		name, valueQuoted := strings.ToLower(match[1]), match[2]
		value, err := strconv.Unquote(valueQuoted)
		if err != nil {
			return nil, err
		}
		params[name] = value
	}

	for _, match := range unquotedParamRegex.FindAllStringSubmatch(gcode, -1) {
		name, value := strings.ToLower(match[1]), match[2]
		if _, exists := params[name]; exists {
			continue
		}
		params[name] = value
	}

	ch := make(chan error)
	go func() {
		defer close(ch)
		ch <- macro.Execute(manager, context, rawParams, objects, params)
	}()
	return ch, nil
}
