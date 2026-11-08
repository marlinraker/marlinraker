package api

import (
	"net/http/httptest"
	"testing"
)

func TestOriginChecker(t *testing.T) {
	checker := newOriginChecker([]string{"*.lan", "http://*.example.com", "*://app.fluidd.xyz", "  "})

	tests := []struct {
		host    string
		origin  string
		allowed bool
	}{
		{"printer.local:7125", "", true},
		{"printer.local:7125", "http://printer.local", true},
		{"printer.local:7125", "http://PRINTER.local:8080", true},
		{"192.168.1.5:7125", "http://192.168.1.5", true},
		{"[::1]:7125", "http://[::1]:8080", true},
		{"printer.local:7125", "http://printer.lan", true},
		{"printer.local:7125", "https://printer.lan:4443", true},
		{"printer.local:7125", "http://my.example.com", true},
		{"printer.local:7125", "https://my.example.com", false},
		{"printer.local:7125", "https://app.fluidd.xyz", true},
		{"printer.local:7125", "https://evil.com", false},
		{"printer.local:7125", "https://printer.lan.evil.com", false},
		{"printer.local:7125", "null", false},
	}

	for _, test := range tests {
		request := httptest.NewRequest("GET", "http://"+test.host+"/server/info", nil)
		if allowed := checker.isAllowed(request, test.origin); allowed != test.allowed {
			t.Errorf("host %q origin %q: allowed = %v, want %v", test.host, test.origin, allowed, test.allowed)
		}
	}

	if !newOriginChecker([]string{"*"}).isAllowed(httptest.NewRequest("GET", "/", nil), "https://evil.com") {
		t.Error(`"*" should allow every origin`)
	}
}

func TestOriginHandler(t *testing.T) {
	handler := newOriginChecker(nil).handler(HttpHandler{})

	request := httptest.NewRequest("POST", "http://printer.local:7125/printer/gcode/script?script=G28", nil)
	request.Header.Set("Origin", "https://evil.com")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != 403 {
		t.Errorf("expected 403 for foreign origin, got %d", recorder.Code)
	}
}
