package printer

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/marlinraker/gcode_store"
	"marlinraker/src/printer/parser"
	"marlinraker/src/shared"
	"strings"
)

type executorContext struct {
	printer         *Printer
	name            string
	commandQueue    []command
	currentCommand  *command
	ready           bool
	subContext      *executorContext
	finishListeners []chan struct{}
}

func newExecutorContext(printer *Printer, name string) *executorContext {
	return &executorContext{
		printer:         printer,
		name:            name,
		commandQueue:    make([]command, 0),
		ready:           true,
		finishListeners: make([]chan struct{}, 0),
	}
}

func (context *executorContext) Name() string {
	return context.name
}

func (context *executorContext) MakeSubContext(name string) (shared.ExecutorContext, error) {
	if context.subContext != nil {
		return nil, errors.New("context already has subcontext")
	}
	context.subContext = newExecutorContext(context.printer, name)
	return context.subContext, nil
}

func (context *executorContext) ReleaseSubContext() {
	context.subContext = nil
}

func (context *executorContext) QueueGcode(gcode string, important bool, silent bool) chan string {

	log.WithField("context", context.name).Debugln("queued " + gcode)

	if !silent {
		gcode_store.LogNow(gcode, gcode_store.Command)
	}

	gcode = parser.CleanGcode(gcode)
	if gcode == "" {
		return nil
	}

	if strings.Contains(gcode, "\n") {
		ch := make(chan string)
		go func() {
			commands := make([]command, 0)
			for _, line := range strings.Split(gcode, "\n") {
				ch := make(chan string)
				commands = append(commands, command{gcode: line, ch: ch})
			}

			if important {
				context.commandQueue = append(commands, context.commandQueue...)
			} else {
				context.commandQueue = append(context.commandQueue, commands...)
			}
			go context.flush()

			responses := make([]string, 0)
			for _, command := range commands {
				if command.ch != nil {
					responses = append(responses, <-command.ch)
				}
			}
			ch <- strings.Join(responses, "\n")
		}()
		return ch
	}

	if context.printer.checkEmergencyCommand(gcode) {
		return nil
	}

	ch := make(chan string)
	cmd := command{gcode: gcode, ch: ch}
	if important {
		context.commandQueue = append([]command{cmd}, context.commandQueue...)
	} else {
		context.commandQueue = append(context.commandQueue, cmd)
	}
	go context.flush()
	return ch
}

func (context *executorContext) CommandFinished() chan struct{} {
	ch := make(chan struct{})
	context.finishListeners = append(context.finishListeners, ch)
	return ch
}

func (context *executorContext) Ready() bool {
	return context.ready && len(context.commandQueue) == 0
}

func (context *executorContext) readLine(line string) {

	if context.subContext != nil {
		context.subContext.readLine(line)
		return
	}

	if context != nil {
		if context.currentCommand.response.Len() > 0 {
			context.currentCommand.response.WriteByte('\n')
		}
		context.currentCommand.response.WriteString(line)
	}

	if strings.HasPrefix(line, "ok") {
		context.finishCommand()
		return
	}
}

func (context *executorContext) finishCommand() {
	if context.currentCommand != nil {
		context.currentCommand.ch <- context.currentCommand.response.String()
		close(context.currentCommand.ch)
		context.printer.handleRequestLine(context.currentCommand.gcode)
		context.currentCommand = nil
	}
	context.ready = true

	if len(context.finishListeners) > 0 {
		for _, ch := range context.finishListeners {
			close(ch)
		}
		context.finishListeners = make([]chan struct{}, 0)
	}
	go context.flush()
}

func (context *executorContext) flush() {

	if !context.ready || len(context.commandQueue) == 0 {
		return
	}

	context.ready = false
	context.currentCommand = &context.commandQueue[0]
	context.commandQueue = context.commandQueue[1:]

	if errCh := context.printer.MacroManager.TryCommand(context, context.currentCommand.gcode); errCh != nil {

		log.WithField("context", context.name).Debugln("macro: " + context.currentCommand.gcode)
		go func() {
			if err := <-errCh; err != nil {
				message := "!! Error: " + err.Error()
				if err := context.printer.Respond(message); err != nil {
					log.Error(err)
				}
			}
			context.currentCommand.response.WriteString("ok")
			context.finishCommand()
		}()
		return
	}

	log.WithField("context", context.name).
		WithField("port", context.printer.path).
		Debugln("write: " + string(context.currentCommand.gcode))

	_, err := context.printer.port.Write([]byte(context.currentCommand.gcode + "\n"))
	if err != nil {
		log.Error(err)
	}
}

func (context *executorContext) resetCommandQueue() {
	context.commandQueue = make([]command, 0)
	context.currentCommand = nil
	context.ready = true
}
