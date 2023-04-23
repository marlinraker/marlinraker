package parser

import (
	"errors"
	"regexp"
)

type PrinterInfo struct {
	MachineType  string
	FirmwareName string
}

var (
	machineTypeRegex  = regexp.MustCompile(`(?m)MACHINE_TYPE:(.*?)(?: [A-Z_]+:|$)`)
	firmwareNameRegex = regexp.MustCompile(`(?m)FIRMWARE_NAME:(.*?)(?: [A-Z_]+:|$)`)
	capabilityRegex   = regexp.MustCompile(`(?m)Cap:([A-Z_]+):([01])`)
)

func ParseM115(response string) (PrinterInfo, map[string]bool, error) {
	machineTypeResult := machineTypeRegex.FindStringSubmatch(response)
	firmwareNameResult := firmwareNameRegex.FindStringSubmatch(response)
	printerInfo := PrinterInfo{}

	if machineTypeResult == nil || firmwareNameResult == nil {
		return printerInfo, nil, errors.New("invalid response")
	}

	printerInfo.MachineType = machineTypeResult[1]
	printerInfo.FirmwareName = firmwareNameResult[1]

	capabilities := make(map[string]bool)
	for _, match := range capabilityRegex.FindAllStringSubmatch(response, -1) {
		name, enabled := match[1], match[2]
		capabilities[name] = enabled == "1"
	}

	return printerInfo, capabilities, nil
}
