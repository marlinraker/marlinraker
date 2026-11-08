package macros

import "github.com/marlinraker/marlinraker/internal/shared"

type cancelPrintMacro struct{}

func (cancelPrintMacro) Description() string {
	return "Cancel the current print"
}

func (cancelPrintMacro) Execute(manager *MacroManager, context shared.ExecutorContext, _ []string, _ Objects, _ Params) error {
	return manager.printer.GetPrintManager().Cancel(context)
}
