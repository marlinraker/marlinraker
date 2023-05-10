package shared

type Printer interface {
	QueueGcode(gcodeRaw string, important bool, silent bool) chan string
	Respond(message string) error
	GetPrintManager() PrintManager
}

type PrintManager interface {
	SelectFile(fileName string) error
	Start() error
	Pause() error
	Resume() error
	Cancel() error
	Reset() error
}
