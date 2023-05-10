package macros

type pauseMacro struct{}

func (pauseMacro) Description() string {
	return "Pauses the current print"
}

func (pauseMacro) Execute(manager *MacroManager, _ []string, _ Objects, _ Params) error {
	return manager.printer.GetPrintManager().Pause()
}
