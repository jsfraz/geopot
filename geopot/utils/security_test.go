package utils

import (
	"strings"
	"testing"
	"time"
)

func TestIsAllowedOrigin(t *testing.T) {
	singleton := GetSingleton()
	original := singleton.Config
	t.Cleanup(func() { singleton.Config = original })

	// Release mode, no allowlist: same-origin only.
	singleton.Config = Config{GinMode: "release"}
	cases := []struct {
		name     string
		origin   string
		host     string
		expected bool
	}{
		{"no origin", "", "geopot.example.com", true},
		{"same origin", "https://geopot.example.com", "geopot.example.com", true},
		{"cross origin", "https://evil.example.com", "geopot.example.com", false},
		{"malformed", "not-a-url", "geopot.example.com", false},
	}
	for _, tc := range cases {
		if got := IsAllowedOrigin(tc.origin, tc.host); got != tc.expected {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.expected)
		}
	}

	// Explicit allowlist.
	singleton.Config = Config{GinMode: "release", AllowedOrigins: []string{"https://allowed.example.com"}}
	if !IsAllowedOrigin("https://allowed.example.com", "geopot.example.com") {
		t.Error("configured origin should be allowed")
	}
	if IsAllowedOrigin("https://evil.example.com", "geopot.example.com") {
		t.Error("unlisted origin should be rejected")
	}

	// Debug mode allows localhost dev server origins.
	singleton.Config = Config{GinMode: "debug"}
	if !IsAllowedOrigin("http://localhost:5173", "localhost:8080") {
		t.Error("localhost origin should be allowed in debug mode")
	}
	if IsAllowedOrigin("https://evil.example.com", "localhost:8080") {
		t.Error("remote origin should stay blocked in debug mode")
	}
}

func TestRateLimiter(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := NewRateLimiter(2)
	limiter.window = time.Minute
	limiter.now = func() time.Time { return now }

	if !limiter.Allow("1.2.3.4") || !limiter.Allow("1.2.3.4") {
		t.Fatal("first two requests should be allowed")
	}
	if limiter.Allow("1.2.3.4") {
		t.Fatal("third request should be rate limited")
	}
	if !limiter.Allow("5.6.7.8") {
		t.Fatal("different key should not be limited")
	}

	now = now.Add(time.Minute + time.Second)
	if !limiter.Allow("1.2.3.4") {
		t.Fatal("window should reset")
	}
}

func TestSanitizeCredential(t *testing.T) {
	if got := SanitizeCredential("root\nInjected log line"); strings.ContainsAny(got, "\n\r") {
		t.Errorf("control characters were not stripped: %q", got)
	}
	if got := SanitizeCredential(strings.Repeat("a", MaxCredentialLength+10)); len(got) != MaxCredentialLength {
		t.Errorf("long value was not truncated: %d", len(got))
	}
	if got := SanitizeCredential("root"); got != "root" {
		t.Errorf("normal value changed: %q", got)
	}
}
