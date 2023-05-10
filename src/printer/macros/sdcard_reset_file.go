package macros

type sdcardResetFileMacro struct{}

func (sdcardResetFileMacro) Description() string {
	return "Clears a loaded SD File. Stops the print if necessary."
}

func (sdcardResetFileMacro) Execute(manager *MacroManager, _ []string, _ Objects, _ Params) error {
	return manager.printer.GetPrintManager().Reset()
}
