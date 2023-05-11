package macros

import "marlinraker/src/shared"

type restoreGcodeState struct {
}

func (restoreGcodeState) Description() string {
	return "Restore a previously saved G-Code state"
}

func (restoreGcodeState) Execute(manager *MacroManager, context shared.ExecutorContext, _ []string, _ Objects, params Params) error {
	name, exists := params["NAME"]
	if !exists {
		name = "default"
	}
	return manager.printer.RestoreGcodeState(context, name)
}
