package constants

import (
	"strconv"
	"strings"
)

// Version is the Marlinraker version. It is overridden at build time with
//
//	-ldflags "-X github.com/marlinraker/marlinraker/internal/constants.Version=1.2.3"
//
// It must stay a plain string literal for the -X linker flag to take effect.
var Version = "0.0.0-dev"

var (
	// ApiVersionString is the "major.minor.patch" part of Version.
	ApiVersionString = apiVersionString(Version)
	ApiVersion       = parseApiVersion(ApiVersionString)
)

// apiVersionString strips a leading "v" and any pre-release or build suffix
// ("v1.2.3-dev+abc" -> "1.2.3").
func apiVersionString(version string) string {
	version = strings.TrimPrefix(version, "v")
	if idx := strings.IndexAny(version, "-+"); idx != -1 {
		version = version[:idx]
	}
	return version
}

func parseApiVersion(version string) [3]int {
	var result [3]int
	for i, part := range strings.SplitN(version, ".", 3) {
		n, err := strconv.Atoi(part)
		if err != nil {
			break
		}
		result[i] = n
	}
	return result
}
