package shared

type Printer interface {
	QueueGcode(gcodeRaw string, important bool, silent bool) (chan string, error)
}
