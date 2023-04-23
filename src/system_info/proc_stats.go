package system_info

import (
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"marlinraker-go/src/api/notification"
	"marlinraker-go/src/marlinraker/connections"
	"sync"
	"time"
)

type snapshot struct {
	Time     float64 `json:"time"`
	CpuUsage float64 `json:"cpu_usage"`
	Memory   int64   `json:"memory"`
	MemUnits string  `json:"mem_units"`
}

type ProcStats struct {
	MoonrakerStats       []snapshot     `json:"moonraker_stats"`
	ThrottledState       throttledState `json:"throttled_state"`
	CpuTemp              float64        `json:"cpu_temp"`
	Network              networkStats   `json:"network"`
	SystemCpuUsage       cpuUsage       `json:"system_cpu_usage"`
	SystemUptime         float64        `json:"system_uptime"`
	WebsocketConnections int            `json:"websocket_connections"`
}

type ProcStat struct {
	MoonrakerStats       snapshot     `json:"moonraker_stats"`
	CpuTemp              float64      `json:"cpu_temp"`
	Network              networkStats `json:"network"`
	SystemCpuUsage       cpuUsage     `json:"system_cpu_usage"`
	WebsocketConnections int          `json:"websocket_connections"`
}

var (
	lastTimes        cpuTimes
	lastNetworkStats *timedNetworkStats
	statsMutex       = &sync.RWMutex{}
	stats            = &ProcStats{}
)

func Run() {
	var err error
	lastTimes, err = getCpuTimes()
	if err != nil {
		log.Error(err)
		return
	}

	lastNetworkStats, err = getNetworkStats(nil)
	if err != nil {
		log.Error(err)
		return
	}

	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		err = takeSnapshot()
		if err != nil {
			log.Error(err)
			break
		}
	}
}

func GetStats() *ProcStats {
	statsMutex.RLock()
	defer statsMutex.RUnlock()
	return stats
}

func takeSnapshot() error {

	statsMutex.Lock()
	defer statsMutex.Unlock()

	now := float64(time.Now().UnixMilli()) / 1000.0

	currentTimes, err := getCpuTimes()
	if err != nil {
		return err
	}

	usedMem, memUnits, err := getUsedMem()
	if err != nil {
		return err
	}

	stats.SystemCpuUsage = getCpuUsage(lastTimes, currentTimes)
	lastTimes = currentTimes
	avgCpuUsage := stats.SystemCpuUsage["cpu"]

	stats.MoonrakerStats = append(stats.MoonrakerStats, snapshot{
		Time:     now,
		CpuUsage: avgCpuUsage,
		Memory:   usedMem,
		MemUnits: memUnits,
	})

	if len(stats.MoonrakerStats) > 30 {
		stats.MoonrakerStats = stats.MoonrakerStats[1:]
	}

	stats.CpuTemp, err = getCpuTemp()
	if err != nil {
		return err
	}

	networkStats, err := getNetworkStats(lastNetworkStats)
	if err != nil {
		return err
	}
	lastNetworkStats = networkStats
	stats.Network = networkStats.Stats

	stats.SystemUptime, err = getUptime()
	if err != nil {
		return err
	}

	stats.WebsocketConnections = len(connections.GetConnections())

	throttledState, _ := getThrottledState()
	if stats.ThrottledState.Bits != throttledState.Bits {
		err = notification.Publish(notification.New("notify_proc_stat_update", []any{throttledState}))
		if err != nil {
			return err
		}
	}
	stats.ThrottledState = throttledState

	notify := notification.New("notify_proc_stat_update", []any{ProcStat{
		MoonrakerStats:       lo.Must(lo.Last(stats.MoonrakerStats)),
		CpuTemp:              stats.CpuTemp,
		Network:              stats.Network,
		SystemCpuUsage:       stats.SystemCpuUsage,
		WebsocketConnections: stats.WebsocketConnections,
	}})

	return notification.Publish(notify)
}
