package shared

type Printer interface {
	Respond(message string) error
	GetPrintManager() PrintManager
	SaveGcodeState(name string)
	RestoreGcodeState(context ExecutorContext, name string) error
	MainExecutorContext() ExecutorContext
	EmergencyStop() error
}

type PrintManager interface {
	SelectFile(fileName string) error
	Start(context ExecutorContext) error
	Pause(context ExecutorContext) error
	Resume(context ExecutorContext) error
	Cancel(context ExecutorContext) error
	Reset(context ExecutorContext) error
}

type ExecutorContext interface {
	Name() string
	QueueGcode(gcodeRaw string, silent bool) chan string
	MakeSubContext(name string) (ExecutorContext, error)
	ReleaseSubContext()
	Pending() chan struct{}
}
