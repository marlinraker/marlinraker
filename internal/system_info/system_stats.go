package system_info

import (
	"fmt"

	"github.com/marlinraker/marlinraker/internal/printer_objects"
	"github.com/marlinraker/marlinraker/internal/system_info/procfs"
)

type systemStatsObject struct{}

func (systemStatsObject) Query() (printer_objects.QueryResult, error) {
	loadAvg, err := procfs.GetLoadAvg()
	if err != nil {
		return nil, fmt.Errorf("failed to get load avg: %w", err)
	}

	processTime, err := procfs.GetProcessTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get process time: %w", err)
	}

	memAvail, err := procfs.GetMemAvail()
	if err != nil {
		return nil, fmt.Errorf("failed to get available memory: %w", err)
	}

	return printer_objects.QueryResult{
		"sysload":  loadAvg[0],
		"cputime":  processTime,
		"memavail": memAvail,
	}, nil
}
