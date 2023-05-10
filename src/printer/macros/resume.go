package macros

type resumeMacro struct{}

func (resumeMacro) Description() string {
	return "Resumes the print from a pause"
}

func (resumeMacro) Execute(manager *MacroManager, _ []string, _ Objects, _ Params) error {
	return manager.printer.GetPrintManager().Resume()
}
