package parser

import "regexp"

var emergencyPrusaRegex = regexp.MustCompile(`^M112(?:\s|$)`)
var emergencyRegex = regexp.MustCompile(`^M(?:112|108|410|876)(?:\s|$)`)

func IsEmergencyCommand(gcode string, isPrusa bool) bool {
	if isPrusa {
		return emergencyPrusaRegex.MatchString(gcode)
	} else {
		return emergencyRegex.MatchString(gcode)
	}
}
