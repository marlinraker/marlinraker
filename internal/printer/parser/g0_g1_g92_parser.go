package parser

import (
	"regexp"
	"strconv"
)

var (
	coordRegex = regexp.MustCompile(` ([XYZEF])([+-]?[0-9.]+)`)
)

func ParseG0G1G92(request string) (map[string]float64, error) {
	coords := make(map[string]float64)
	for _, match := range coordRegex.FindAllStringSubmatch(request, -1) {
		value, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			return coords, err
		}
		coords[match[1]] = value
	}
	return coords, nil
}
