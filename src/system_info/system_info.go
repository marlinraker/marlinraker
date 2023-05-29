package system_info

import (
	"github.com/samber/lo"
	"marlinraker/src/service"
	"marlinraker/src/system_info/procfs"
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
		return nil, err
	}

	sdInfo, _ := procfs.GetSdInfo()

	distribution, err := procfs.GetDistribution()
	if err != nil {
		return nil, err
	}

	network, err := procfs.GetNetwork()
	if err != nil {
		return nil, err
	}

	state, err := service.GetServiceState()
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		CpuInfo:           cpuInfo,
		SdInfo:            sdInfo,
		Distribution:      distribution,
		AvailableServices: lo.Keys(state),
		InstanceIds:       map[string]string{},
		ServiceState:      state,
		Virtualization:    &virtualization{"none", "none"},
		Network:           network,
	}, nil
}
