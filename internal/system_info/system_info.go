package system_info

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/marlinraker/marlinraker/internal/service"
	"github.com/marlinraker/marlinraker/internal/system_info/procfs"
)

type virtualization struct {
	VirtType       string `json:"virt_type"`
	VirtIdentifier string `json:"virt_identifier"`
}

type SystemInfo struct {
	CpuInfo           *procfs.CpuInfo              `json:"cpu_info"`
	SdInfo            *procfs.SdInfo               `json:"sd_info"`
	Distribution      *procfs.Distribution         `json:"distribution"`
	AvailableServices []string                     `json:"available_services"`
	InstanceIds       map[string]string            `json:"instance_ids"`
	ServiceState      map[string]service.UnitState `json:"service_state"`
	Virtualization    *virtualization              `json:"virtualization"`
	Network           map[string]procfs.Network    `json:"network"`
}

func GetSystemInfo() (*SystemInfo, error) {

	cpuInfo, err := procfs.GetCpuInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu info: %w", err)
	}

	sdInfo, _ := procfs.GetSdInfo()

	distribution, err := procfs.GetDistribution()
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution info: %w", err)
	}

	network, err := procfs.GetNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %w", err)
	}

	info := &SystemInfo{
		CpuInfo:           cpuInfo,
		SdInfo:            sdInfo,
		Distribution:      distribution,
		AvailableServices: []string{},
		InstanceIds:       map[string]string{},
		ServiceState:      map[string]service.UnitState{},
		Virtualization:    &virtualization{"none", "none"},
		Network:           network,
	}

	state, err := service.GetServiceState()
	if err == nil {
		info.AvailableServices = lo.Keys(state)
		info.ServiceState = state
	}
	return info, nil
}
