package macros

import "marlinraker/src/shared"

type resumeMacro struct{}

func (resumeMacro) Description() string {
	return "Resumes the print from a pause"
}

func (resumeMacro) Execute(manager *MacroManager, context shared.ExecutorContext, _ []string, _ Objects, _ Params) error {
	return manager.printer.GetPrintManager().Resume(context)
}
