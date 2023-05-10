package macros

type cancelPrintMacro struct{}

func (cancelPrintMacro) Description() string {
	return "Cancel the current print"
}

func (cancelPrintMacro) Execute(manager *MacroManager, _ []string, _ Objects, _ Params) error {
	return manager.printer.GetPrintManager().Cancel()
}
