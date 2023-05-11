package macros

import "marlinraker/src/shared"

type saveGcodeState struct {
}

func (saveGcodeState) Description() string {
	return "Save G-Code coordinate state"
}

func (saveGcodeState) Execute(manager *MacroManager, _ shared.ExecutorContext, _ []string, _ Objects, params Params) error {
	name, exists := params["NAME"]
	if !exists {
		name = "default"
	}
	manager.printer.SaveGcodeState(name)
	return nil
}
