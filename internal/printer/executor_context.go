package printer

import (
	stdcontext "context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/marlinraker/gcode_store"
	"github.com/marlinraker/marlinraker/internal/printer/parser"
	"github.com/marlinraker/marlinraker/internal/shared"
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
	closeOnce       sync.Once
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

// close stops the executor. Commands that are queued or in flight are released
// with an empty response and commands queued afterward complete immediately.
func (context *executorContext) close() {
	context.closeOnce.Do(func() {
		context.mu.Lock()
		defer context.mu.Unlock()
		close(context.closeCh)
	})
}

func (context *executorContext) isClosed() bool {
	select {
	case <-context.closeCh:
		return true
	default:
		return false
	}
}

// drain releases all commands that are still waiting in the queue.
func (context *executorContext) drain() {
	for {
		select {
		case cmd := <-context.commandCh:
			close(cmd.ch)
			context.pending.Done()
		default:
			return
		}
	}
}

// await waits for the response of a queued command. It gives up when ctx is
// done or when the executor context is closed, so callers never block forever
// on a printer that stopped responding.
func (context *executorContext) await(ctx stdcontext.Context, ch <-chan string) (string, error) {
	select {
	case response := <-ch:
		return response, nil
	case <-context.closeCh:
		return "", errors.New("executor context closed")
	case <-ctx.Done():
		return "", ctx.Err()
	}
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
			context.drain()
			return

		case cmd := <-context.commandCh:
			context.flush(cmd)

			var response string
			select {
			case response = <-context.responseCh:
			case <-context.closeCh:
				log.WithField("context", context.name).Debugln("end")
				close(cmd.ch)
				context.pending.Done()
				context.drain()
				return
			}
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

	log.WithField("context", context.name).Debugf("queued %s", gcode)

	if !silent {
		gcode_store.LogNow(gcode, gcode_store.Command)
	}

	gcode = parser.CleanGcode(gcode)
	if gcode == "" || context.isClosed() {
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

		responseCh := make(chan string, 1)
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

	log.WithField("context", context.name).Debugf("read: %s", line)

	if context.responseBuilder.Len() > 0 {
		context.responseBuilder.WriteByte('\n')
	}
	context.responseBuilder.WriteString(line)

	if strings.HasPrefix(line, "ok") {
		select {
		case context.responseCh <- context.responseBuilder.String():
		case <-context.closeCh:
		}
	}
}

func (context *executorContext) flush(cmd command) {

	if macro, name, exists := context.printer.MacroManager.GetMacro(cmd.gcode); exists {
		log.WithField("context", context.name).Debugf("macro: %s", cmd.gcode)

		subContext, err := context.MakeSubContext(fmt.Sprintf("%s/%s", context.name, name))
		if err != nil {
			log.Errorf("Could not create subcontext: %v", err)
			return
		}

		ch, err := context.printer.MacroManager.ExecuteMacro(macro, subContext, cmd.gcode)
		if err != nil {
			err = <-ch
		}
		if err != nil {
			message := fmt.Sprintf("!! Error: %s", err)
			if err = context.printer.Respond(message); err != nil {
				log.Errorf("Failed to send response: %v", err)
			}
		}
		context.ReleaseSubContext()

		go func() {
			select {
			case context.responseCh <- "ok":
			case <-context.closeCh:
			}
		}()
		return
	}

	log.WithField("context", context.name).
		WithField("port", context.printer.path).
		Debugf("write: %s\n", cmd.gcode)

	_, err := fmt.Fprintln(context.printer.port, cmd.gcode)
	if err != nil {
		log.Errorf("Failed writing to printer port: %v", err)
	}
}
