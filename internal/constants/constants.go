package constants

// Version is the Marlinraker version. It is overridden at build time with
//
//	-ldflags "-X github.com/marlinraker/marlinraker/internal/constants.Version=1.2.3"
//
// It must stay a plain string literal for the -X linker flag to take effect.
var Version = "0.0.0-dev"

// The Moonraker release and API version whose API Marlinraker implements.
// Clients such as Mainsail and Fluidd compare these against their minimum
// supported versions and enable features based on them, so they describe the
// emulated Moonraker, not Marlinraker itself.
const (
	EmulatedMoonrakerVersion = "v0.9.3"
	ApiVersionString         = "1.4.0"
)

var ApiVersion = [3]int{1, 4, 0}

// MoonrakerVersion is reported as moonraker_version. It starts with the
// emulated Moonraker version so version checks in clients pass, followed by
// the Marlinraker version.
func MoonrakerVersion() string {
	return EmulatedMoonrakerVersion + "-marlinraker-" + Version
}
