package printer

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"marlinraker/src/marlinraker/gcode_store"
	"marlinraker/src/printer/parser"
	"marlinraker/src/shared"
	"marlinraker/src/util"
	"strings"
	"sync"
	"sync/atomic"
)

type command struct {
	gcode string
	ch    chan string
}

type executorContext struct {
	printer         *Printer
	name            string
	subContext      atomic.Pointer[executorContext]
	closeCh         chan struct{}
	commandCh       chan command
	responseCh      chan string
	responseBuilder strings.Builder
	pending         *sync.WaitGroup
	mu              *sync.Mutex
}

func newExecutorContext(printer *Printer, name string) *executorContext {
	context := &executorContext{
		printer:    printer,
		name:       name,
		closeCh:    make(chan struct{}),
		commandCh:  make(chan command, 128),
		responseCh: make(chan string),
		pending:    &sync.WaitGroup{},
		mu:         &sync.Mutex{},
	}
	go context.work()
	return context
}

func (context *executorContext) close() {
	close(context.closeCh)
}

func (context *executorContext) Name() string {
	return context.name
}

func (context *executorContext) Pending() chan struct{} {
	ch := make(chan struct{})
	go func() {
		context.pending.Wait()
		close(ch)
	}()
	return ch
}

func (context *executorContext) MakeSubContext(name string) (shared.ExecutorContext, error) {
	if ctx := context.subContext.Load(); ctx != nil {
		return nil, errors.New("context already has subcontext")
	}
	subContext := newExecutorContext(context.printer, name)
	context.subContext.Store(subContext)
	return subContext, nil
}

func (context *executorContext) ReleaseSubContext() {
	ctx := context.subContext.Load()
	context.subContext.Store(nil)
	ctx.close()
}

func (context *executorContext) work() {
	log.WithField("context", context.name).Debugln("begin")
	for {
		select {
		case <-context.closeCh:
			log.WithField("context", context.name).Debugln("end")
			return

		case cmd := <-context.commandCh:
			context.flush(cmd)

			response := <-context.responseCh
			context.responseBuilder = strings.Builder{}
			select {
			case cmd.ch <- response:
			default:
			}
			close(cmd.ch)
			context.printer.handleRequestLine(cmd.gcode)
			context.pending.Done()
		}
	}
}

func (context *executorContext) QueueGcode(gcode string, silent bool) chan string {

	context.mu.Lock()
	defer context.mu.Unlock()

	log.WithField("context", context.name).Debugln("queued " + gcode)

	if !silent {
		gcode_store.LogNow(gcode, gcode_store.Command)
	}

	gcode = parser.CleanGcode(gcode)
	if gcode == "" {
		ch := make(chan string)
		close(ch)
		return ch
	}

	if strings.Contains(gcode, "\n") {
		chans := make([]chan string, 0)
		lines := strings.Split(gcode, "\n")
		context.pending.Add(len(lines))

		for _, line := range lines {
			ch := make(chan string)
			chans = append(chans, ch)
			cmd := command{gcode: line, ch: ch}
			context.commandCh <- cmd
		}

		responseCh := make(chan string)
		go func() {
			response := make([]string, 0)
			for _, ch := range chans {
				response = append(response, <-ch)
			}
			responseCh <- strings.Join(response, "\n")
		}()
		return responseCh
	}

	ch := make(chan string)
	if context.printer.executeEmergencyCommand(gcode) {
		close(ch)
		return ch
	}

	cmd := command{gcode: gcode, ch: ch}
	context.pending.Add(1)
	context.commandCh <- cmd
	return ch
}

func (context *executorContext) readLine(line string) {

	if subContext := context.subContext.Load(); subContext != nil {
		subContext.readLine(line)
		return
	}

	log.WithField("context", context.name).Debugln("read: " + line)

	if context.responseBuilder.Len() > 0 {
		context.responseBuilder.WriteByte('\n')
	}
	context.responseBuilder.WriteString(line)

	if strings.HasPrefix(line, "ok") {
		context.responseCh <- context.responseBuilder.String()
	}
}

func (context *executorContext) flush(cmd command) {

	if macro, name, exists := context.printer.MacroManager.GetMacro(cmd.gcode); exists {
		log.WithField("context", context.name).Debugln("macro: " + cmd.gcode)

		subContext, err := context.MakeSubContext(context.Name() + "/" + name)
		if err != nil {
			panic(err)
		}

		if err := <-context.printer.MacroManager.ExecuteMacro(macro, subContext, cmd.gcode); err != nil {
			message := "!! Error: " + err.Error()
			if err := context.printer.Respond(message); err != nil {
				util.LogError(err)
			}
		}
		context.ReleaseSubContext()

		go func() {
			context.responseCh <- "ok"
		}()
		return
	}

	log.WithField("context", context.name).
		WithField("port", context.printer.path).
		Debugln("write: " + string(cmd.gcode))

	_, err := context.printer.port.Write([]byte(cmd.gcode + "\n"))
	if err != nil {
		util.LogError(err)
	}
}
