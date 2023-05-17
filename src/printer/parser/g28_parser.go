package parser

import "strings"

func ParseG28(line string) [3]bool {

	var (
		homedAxes     [3]bool
		axisSpecified bool
	)

	for i, char := range []byte("XYZ") {
		if strings.IndexByte(line, char) > 3 {
			homedAxes[i] = true
			axisSpecified = true
		}
	}

	if !axisSpecified {
		for i := 0; i < 3; i++ {
			homedAxes[i] = true
		}
	}
	return homedAxes
}
