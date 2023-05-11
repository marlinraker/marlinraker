package macros

import "marlinraker/src/shared"

type sdcardPrintFileMacro struct{}

func (sdcardPrintFileMacro) Description() string {
	return "Loads a SD file and starts the print. May include files in subdirectories."
}

func (sdcardPrintFileMacro) Execute(manager *MacroManager, context shared.ExecutorContext, _ []string, _ Objects, params Params) error {
	fileName, err := params.RequireString("filename")
	if err != nil {
		return err
	}

	printManager := manager.printer.GetPrintManager()
	if err := printManager.SelectFile(fileName); err != nil {
		return err
	}

	if err := printManager.Start(context); err != nil {
		return err
	}
	return nil
}
