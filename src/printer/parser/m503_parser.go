package parser

import (
	"errors"
	"regexp"
	"strconv"
)

type PrinterLimits struct {
	MaxAccel    [3]float32
	MaxFeedrate [3]float32
}

var (
	accelerationRegex = regexp.MustCompile(`(?m)^echo: *M201 X([0-9.]+) Y([0-9.]+) Z([0-9.]+).*$`)
	feedrateRegex     = regexp.MustCompile(`(?m)^echo: *M203 X([0-9.]+) Y([0-9.]+) Z([0-9.]+).*$`)
)

func ParseM503(response string) (PrinterLimits, error) {

	limits := PrinterLimits{}

	accel := accelerationRegex.FindStringSubmatch(response)
	if accel == nil {
		return limits, errors.New("invalid response")
	}

	accelX, _ := strconv.ParseFloat(accel[1], 32)
	accelY, _ := strconv.ParseFloat(accel[2], 32)
	accelZ, _ := strconv.ParseFloat(accel[3], 32)
	limits.MaxAccel = [3]float32{float32(accelX), float32(accelY), float32(accelZ)}

	feedrate := feedrateRegex.FindStringSubmatch(response)
	if feedrate == nil {
		return limits, errors.New("invalid response")
	}

	feedrateX, _ := strconv.ParseFloat(feedrate[1], 32)
	feedrateY, _ := strconv.ParseFloat(feedrate[2], 32)
	feedrateZ, _ := strconv.ParseFloat(feedrate[3], 32)
	limits.MaxFeedrate = [3]float32{float32(feedrateX), float32(feedrateY), float32(feedrateZ)}

	return limits, nil
}
