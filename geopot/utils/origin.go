package utils

import (
	"net/url"
	"strings"
)

// Checks whether a browser Origin header is allowed to access the API or
// WebSocket endpoint.
//
// When ALLOWED_ORIGINS is configured, the origin must match one of the entries
// (either a full origin such as https://example.com or a bare host). When the
// list is empty, only same-origin requests are accepted. Requests without an
// Origin header (non-browser clients, curl, SSH tooling) are allowed.
//
//	@param origin
//	@param requestHost
//	@return bool
func IsAllowedOrigin(origin string, requestHost string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return true
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	// Local development: the Vite dev server runs on a different port.
	if GetSingleton().Config.IsDebug() {
		switch parsed.Hostname() {
		case "localhost", "127.0.0.1", "::1":
			return true
		}
	}

	allowedOrigins := GetSingleton().Config.AllowedOrigins
	if len(allowedOrigins) > 0 {
		normalizedOrigin := strings.TrimSuffix(origin, "/")
		for _, allowed := range allowedOrigins {
			allowed = strings.TrimSuffix(strings.TrimSpace(allowed), "/")
			if allowed == "" {
				continue
			}
			// Full origin (scheme + host) comparison.
			if strings.Contains(allowed, "://") {
				if strings.EqualFold(allowed, normalizedOrigin) {
					return true
				}
				continue
			}
			// Bare host comparison (ignores scheme).
			if strings.EqualFold(allowed, parsed.Host) {
				return true
			}
		}
		return false
	}

	// Default: same-origin only.
	return strings.EqualFold(parsed.Host, requestHost)
}
