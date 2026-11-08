package parser

import "regexp"

var actionRegex = regexp.MustCompile(`^//\s*action:\s*([A-Za-z]+)$`)

func ParseAction(response string) string {
	if match := actionRegex.FindStringSubmatch(response); match != nil {
		return match[1]
	}
	return ""
}
