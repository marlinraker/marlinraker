package macros

import (
	"github.com/Masterminds/sprig/v3"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/shared"
	"strings"
	"text/template"
)

type customMacro struct {
	description string
	tmpl        *template.Template
}

func newCustomMacro(name string, description string, content string) (*customMacro, error) {
	tmpl, err := template.New(name).Funcs(sprig.FuncMap()).Parse(content)
	if err != nil {
		return nil, err
	}
	return &customMacro{description: description, tmpl: tmpl}, nil
}

func (macro *customMacro) Description() string {
	return macro.description
}

type macroContext struct {
	manager   *MacroManager
	Printer   Objects
	Params    Params
	RawParams []string
}

func (context macroContext) ActionRespondInfo(message string) string {
	if err := context.manager.printer.Respond("// " + message); err != nil {
		log.Error(err)
	}
	return ""
}

func (context macroContext) ActionRaiseError(message string) string {
	if err := context.manager.printer.Respond("!! " + message); err != nil {
		log.Error(err)
	}
	return ""
}

func (macro *customMacro) Execute(manager *MacroManager, context shared.ExecutorContext, rawParams []string, objects Objects, params Params) error {

	ctx := macroContext{
		manager:   manager,
		Printer:   objects,
		Params:    params,
		RawParams: rawParams,
	}

	builder := strings.Builder{}
	if err := macro.tmpl.Execute(&builder, ctx); err != nil {
		return err
	}
	gcode := builder.String()
	<-context.QueueGcode(gcode, true)
	return nil
}
