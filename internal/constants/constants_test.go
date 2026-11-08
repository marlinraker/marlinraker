package constants

import (
	"fmt"
	"testing"
)

func TestApiVersion(t *testing.T) {
	if got := fmt.Sprintf("%d.%d.%d", ApiVersion[0], ApiVersion[1], ApiVersion[2]); got != ApiVersionString {
		t.Errorf("ApiVersion %s does not match ApiVersionString %s", got, ApiVersionString)
	}
}

func TestMoonrakerVersion(t *testing.T) {
	old := Version
	t.Cleanup(func() { Version = old })
	Version = "1.2.3"
	if got, want := MoonrakerVersion(), "v0.9.3-marlinraker-1.2.3"; got != want {
		t.Errorf("MoonrakerVersion() = %q, want %q", got, want)
	}
}
