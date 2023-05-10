package macros

type renamedMacro struct {
	description string
	original    Macro
}

func (macro renamedMacro) Description() string {
	return macro.description
}

func (macro renamedMacro) Execute(manager *MacroManager, rawParams []string, objects Objects, params Params) error {
	return macro.original.Execute(manager, rawParams, objects, params)
}
