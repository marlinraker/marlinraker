package printer

import (
	"errors"
	"github.com/samber/lo"
	"regexp"
	"strconv"
	"strings"
)

var (
	setHeaterTempRegex = regexp.MustCompile(`(?i)SET_HEATER_TEMPERATURE(?:\s|$)`)
	heaterRegex        = regexp.MustCompile(`(?i)\sHEATER=(\S+)(?:\s|$)`)
	targetRegex        = regexp.MustCompile(`(?i)\sTARGET=(\S+)(?:\s|$)`)
)

func (printer *Printer) tryKlipperCommand(command string) (chan string, error) {

	switch {
	case setHeaterTempRegex.MatchString(command):
		heaterMatch := heaterRegex.FindStringSubmatch(command)
		if heaterMatch == nil {
			return nil, errors.New("missing argument HEATER")
		}
		heater := heaterMatch[1]

		if !lo.Contains(printer.heaters.availableHeaters, heater) {
			return nil, errors.New("cannot find heater \"" + heater + "\"")
		}

		targetMatch := targetRegex.FindStringSubmatch(command)
		if targetMatch == nil {
			return nil, errors.New("missing argument TARGET")
		}
		target, err := strconv.ParseFloat(targetMatch[1], 64)
		if err != nil {
			return nil, err
		}

		switch {
		case strings.HasPrefix(heater, "extruder"):
			idx := heater[8:]
			if idx == "" {
				idx = "0"
			}
			return printer.QueueGcode("M104 T"+idx+" S"+strconv.FormatFloat(target, 'f', 2, 64), false, true)
		}
	}

	return nil, nil
}
