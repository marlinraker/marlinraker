package macros

import "marlinraker/src/shared"

type renamedMacro struct {
	description string
	original    Macro
}

func (macro renamedMacro) Description() string {
	return macro.description
}

func (macro renamedMacro) Execute(manager *MacroManager, context shared.ExecutorContext, rawParams []string, objects Objects, params Params) error {
	return macro.original.Execute(manager, context, rawParams, objects, params)
}
