package parser

import (
	"regexp"
	"strings"
)

var emergencyPrusaRegex = regexp.MustCompile(`^M112(?:\s|$)`)
var emergencyRegex = regexp.MustCompile(`^M(?:112|108|410|876)(?:\s|$)`)

func IsEmergencyCommand(gcode string, isPrusa bool) bool {
	if isPrusa {
		return emergencyPrusaRegex.MatchString(gcode)
	} else {
		return emergencyRegex.MatchString(gcode)
	}
}

func CleanGcode(gcode string) string {
	var lines []string
	for _, line := range strings.Split(gcode, "\n") {
		idx := strings.Index(line, ";")
		if idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
