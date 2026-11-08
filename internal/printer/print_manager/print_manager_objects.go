package print_manager

import "github.com/marlinraker/marlinraker/internal/printer_objects"

type printStatsObject struct {
	manager *PrintManager
}

func (object printStatsObject) Query() (printer_objects.QueryResult, error) {

	var (
		fileName      string
		totalDuration float64
		printDuration float64
	)

	if job := object.manager.currentJob.Load(); job != nil {
		fileName = job.fileName
		totalDuration = job.getTotalTime().Seconds()
		printDuration = job.getPrintTime().Seconds()
	}

	return printer_objects.QueryResult{
		"filename":       fileName,
		"total_duration": totalDuration,
		"print_duration": printDuration,
		"filament_used":  object.manager.getFilamentUsed(),
		"state":          object.manager.state.Load(),
		"message":        "",
	}, nil
}

type virtualSdcardObject struct {
	manager *PrintManager
}

func (object virtualSdcardObject) Query() (printer_objects.QueryResult, error) {

	var (
		filePath string
		progress float64
		size     int64
		position int64
	)

	if job := object.manager.currentJob.Load(); job != nil {
		filePath, progress, size, position =
			job.filePath, job.progress.Load(), job.fileSize, job.position.Load()
	}

	return printer_objects.QueryResult{
		"is_active":     object.manager.state.Load() == "printing",
		"progress":      progress,
		"file_path":     filePath,
		"file_position": position,
		"file_size":     size,
	}, nil
}

type pauseResumeObject struct {
	manager *PrintManager
}

func (object pauseResumeObject) Query() (printer_objects.QueryResult, error) {
	return printer_objects.QueryResult{
		"is_paused": object.manager.state.Load() == "paused",
	}, nil
}
