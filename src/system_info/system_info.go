package system_info

import "marlinraker/src/system_info/procfs"

type serviceState struct {
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
}

type virtualization struct {
	VirtType       string `json:"virt_type"`
	VirtIdentifier string `json:"virt_identifier"`
}

type SystemInfo struct {
	CpuInfo           *procfs.CpuInfo           `json:"cpu_info"`
	SdInfo            *procfs.SdInfo            `json:"sd_info"`
	Distribution      *procfs.Distribution      `json:"distribution"`
	AvailableServices []string                  `json:"available_services"`
	InstanceIds       map[string]string         `json:"instance_ids"`
	ServiceState      map[string]serviceState   `json:"service_state"`
	Virtualization    *virtualization           `json:"virtualization"`
	Network           map[string]procfs.Network `json:"network"`
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

	return &SystemInfo{
		CpuInfo:           cpuInfo,
		SdInfo:            sdInfo,
		Distribution:      distribution,
		AvailableServices: []string{},
		InstanceIds:       map[string]string{},
		ServiceState:      map[string]serviceState{},
		Virtualization:    &virtualization{"none", "none"},
		Network:           network,
	}, nil
}
