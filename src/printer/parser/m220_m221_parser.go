package parser

import (
	"errors"
	"regexp"
	"strconv"
)

var (
	sRegex = regexp.MustCompile(`S([0-9]+)`)
)

func ParseM220M221(request string) (int32, error) {
	if match := sRegex.FindStringSubmatch(request); match != nil {
		s, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return 0, err
		}
		return int32(s), nil
	}
	return 0, errors.New("missing S parameter")
}
