package api

import (
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// originChecker decides which browser origins may talk to the API.
//
// An origin is allowed when
//   - the request has no Origin header (curl, slicers, other non-browser clients),
//   - the origin has the same hostname as the request (e.g. Mainsail served from the same host), or
//   - the origin matches one of the configured cors_domains patterns.
//
// Patterns may contain "*" wildcards. A pattern that contains "://" is matched
// against the full origin (e.g. "http://*.local:8080"); any other pattern is
// matched against the origin's hostname only (e.g. "*.local").
// The pattern "*" allows every origin.
type originChecker struct {
	allowAll bool
	full     []*regexp.Regexp
	hostname []*regexp.Regexp
}

func newOriginChecker(patterns []string) *originChecker {
	checker := &originChecker{}
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		switch {
		case pattern == "":
			continue
		case pattern == "*":
			checker.allowAll = true
		case strings.Contains(pattern, "://"):
			checker.full = append(checker.full, globToRegexp(strings.TrimRight(pattern, "/")))
		default:
			checker.hostname = append(checker.hostname, globToRegexp(pattern))
		}
	}
	return checker
}

func globToRegexp(pattern string) *regexp.Regexp {
	parts := strings.Split(pattern, "*")
	for i, part := range parts {
		parts[i] = regexp.QuoteMeta(part)
	}
	return regexp.MustCompile("^" + strings.Join(parts, ".*") + "$")
}

func (checker *originChecker) isAllowed(request *http.Request, origin string) bool {
	if origin == "" || checker.allowAll {
		return true
	}

	originUrl, err := url.Parse(strings.ToLower(origin))
	if err != nil || originUrl.Host == "" {
		return false
	}

	if originUrl.Hostname() == requestHostname(request) {
		return true
	}

	normalizedOrigin := originUrl.Scheme + "://" + originUrl.Host
	for _, pattern := range checker.full {
		if pattern.MatchString(normalizedOrigin) {
			return true
		}
	}
	for _, pattern := range checker.hostname {
		if pattern.MatchString(originUrl.Hostname()) {
			return true
		}
	}
	return false
}

func requestHostname(request *http.Request) string {
	host := request.Host
	if forwarded := request.Header.Get("X-Forwarded-Host"); forwarded != "" {
		host = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		host = hostname
	}
	return strings.ToLower(strings.Trim(host, "[]"))
}

// handler rejects requests from browser origins that are not allowed, so that
// arbitrary web pages cannot send commands to the printer (CSRF).
func (checker *originChecker) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if !checker.isAllowed(request, origin) {
			log.Warnf("Rejected %s %s from origin %q, add it to cors_domains to allow it", request.Method, request.URL.Path, origin)
			http.Error(writer, "Origin not allowed", http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
