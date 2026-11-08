package printer

import (
	"sync"
	"testing"
	"time"

	"go.bug.st/serial"

	"github.com/marlinraker/marlinraker/internal/config"
	"github.com/marlinraker/marlinraker/internal/printer/macros"
)

// fakePort is a serial port that swallows writes and never answers.
type fakePort struct {
	serial.Port
	mu      sync.Mutex
	written []string
}

func (port *fakePort) Write(p []byte) (int, error) {
	port.mu.Lock()
	defer port.mu.Unlock()
	port.written = append(port.written, string(p))
	return len(p), nil
}

func newTestPrinter() *Printer {
	cfg := &config.Config{}
	printer := &Printer{config: cfg, port: &fakePort{}, path: "fake"}
	printer.MacroManager = macros.NewMacroManager(printer, cfg)
	printer.GcodeState = &GcodeState{}
	return printer
}

func receiveWithin(t *testing.T, ch <-chan string, timeout time.Duration) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatal("timed out waiting for channel")
	}
}

func TestExecutorContextCloseReleasesWaiters(t *testing.T) {
	printer := newTestPrinter()
	ctx := newExecutorContext(printer, "test")

	// The fake port never answers with "ok", so these would block forever.
	inFlight := ctx.QueueGcode("G28", true)
	queued := ctx.QueueGcode("M400", true)

	ctx.close()

	receiveWithin(t, inFlight, time.Second)
	receiveWithin(t, queued, time.Second)
	receiveWithin(t, ctx.QueueGcode("M400", true), time.Second)

	select {
	case <-ctx.Pending():
	case <-time.After(time.Second):
		t.Fatal("pending commands were not released")
	}

	// close must be idempotent
	ctx.close()
}

func TestExecutorContextReadLineAfterClose(t *testing.T) {
	printer := newTestPrinter()
	ctx := newExecutorContext(printer, "test")
	ctx.close()

	done := make(chan struct{})
	go func() {
		ctx.readLine("ok")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("readLine blocked on a closed context")
	}
}

func TestWatcherTimersStop(t *testing.T) {
	position := &positionWatcher{closeCh: make(chan struct{}), ticker: time.NewTicker(time.Hour)}
	temp := &tempWatcher{
		closeCh:       make(chan struct{}),
		ticker:        time.NewTicker(time.Hour),
		heaterObjects: map[string]*heaterObject{},
		sensorObjects: map[string]*temperatureSensorObject{},
	}

	positionDone, tempDone := make(chan struct{}), make(chan struct{})
	go func() { position.runTimer(); close(positionDone) }()
	go func() { temp.runTimer(); close(tempDone) }()

	position.stop()
	position.stop()
	temp.stop()
	temp.stop()

	for _, ch := range []chan struct{}{positionDone, tempDone} {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("watcher timer loop did not exit after stop")
		}
	}
}
