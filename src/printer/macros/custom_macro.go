package macros

import (
	"strings"
	"text/template"
)

type customMacro struct {
	description string
	tmpl        *template.Template
}

func newCustomMacro(name string, description string, content string) (*customMacro, error) {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return nil, err
	}
	return &customMacro{description: description, tmpl: tmpl}, nil
}

func (macro *customMacro) Description() string {
	return macro.description
}

type macroContext struct {
	Printer   Objects
	Params    Params
	RawParams []string
}

func (macro *customMacro) Execute(manager *MacroManager, rawParams []string, objects Objects, params Params) error {
	context := macroContext{objects, params, rawParams}
	builder := strings.Builder{}
	if err := macro.tmpl.Execute(&builder, context); err != nil {
		return err
	}
	gcode := builder.String()
	<-manager.printer.QueueGcode(gcode, false, true)
	return nil
}
