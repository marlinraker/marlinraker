package macros

import (
	"marlinraker/src/shared"
)

type pauseMacro struct{}

func (pauseMacro) Description() string {
	return "Pauses the current print"
}

func (pauseMacro) Execute(manager *MacroManager, context shared.ExecutorContext, _ []string, _ Objects, _ Params) error {
	return manager.printer.GetPrintManager().Pause(context)
}
