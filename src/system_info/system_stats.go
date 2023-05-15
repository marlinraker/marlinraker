package system_info

import (
	"marlinraker/src/printer_objects"
	"marlinraker/src/system_info/procfs"
	"marlinraker/src/util"
)

type systemStatsObject struct{}

func (systemStatsObject) Query() printer_objects.QueryResult {
	loadAvg, err := procfs.GetLoadAvg()
	if err != nil {
		util.LogError(err)
		return nil
	}

	processTime, err := procfs.GetProcessTime()
	if err != nil {
		util.LogError(err)
		return nil
	}

	memAvail, err := procfs.GetMemAvail()
	if err != nil {
		util.LogError(err)
		return nil
	}

	return printer_objects.QueryResult{
		"sysload":  loadAvg[0],
		"cputime":  processTime,
		"memavail": memAvail,
	}
}
