package constants

import "testing"

func TestApiVersion(t *testing.T) {
	tests := []struct {
		version string
		str     string
		parsed  [3]int
	}{
		{"0.0.0-dev", "0.0.0", [3]int{0, 0, 0}},
		{"1.2.3", "1.2.3", [3]int{1, 2, 3}},
		{"v1.2.3", "1.2.3", [3]int{1, 2, 3}},
		{"1.2.4-dev+abc1234", "1.2.4", [3]int{1, 2, 4}},
		{"1.2.3+abc1234", "1.2.3", [3]int{1, 2, 3}},
		{"garbage", "garbage", [3]int{0, 0, 0}},
		{"1.2", "1.2", [3]int{1, 2, 0}},
	}
	for _, test := range tests {
		str := apiVersionString(test.version)
		if str != test.str {
			t.Errorf("apiVersionString(%q) = %q, want %q", test.version, str, test.str)
		}
		if parsed := parseApiVersion(str); parsed != test.parsed {
			t.Errorf("parseApiVersion(%q) = %v, want %v", str, parsed, test.parsed)
		}
	}
}
