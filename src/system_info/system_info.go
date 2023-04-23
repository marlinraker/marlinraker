package system_info

type serviceState struct {
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
}

type virtualization struct {
	VirtType       string `json:"virt_type"`
	VirtIdentifier string `json:"virt_identifier"`
}

type SystemInfo struct {
	CpuInfo           *cpuInfo                `json:"cpu_info"`
	SdInfo            *sdInfo                 `json:"sd_info"`
	Distribution      *distribution           `json:"distribution"`
	AvailableServices []string                `json:"available_services"`
	InstanceIds       map[string]string       `json:"instance_ids"`
	ServiceState      map[string]serviceState `json:"service_state"`
	Virtualization    *virtualization         `json:"virtualization"`
	Network           map[string]network      `json:"network"`
}

func GetSystemInfo() (*SystemInfo, error) {

	cpuInfo, err := getCpuInfo()
	if err != nil {
		return nil, err
	}

	sdInfo, _ := getSdInfo()

	distribution, err := getDistribution()
	if err != nil {
		return nil, err
	}

	network, err := getNetwork()
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
