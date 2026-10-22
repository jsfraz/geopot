package utils

import "testing"

func TestRandomCoordinateStaysInRange(t *testing.T) {
	cases := []struct {
		name     string
		min, max float64
	}{
		{"latitude", -90, 90},
		{"longitude", -180, 180},
		{"single point", 10, 10},
	}

	for _, tc := range cases {
		for i := 0; i < 1000; i++ {
			got := RandomCoordinate(tc.min, tc.max)
			if got < tc.min || got > tc.max {
				t.Fatalf("%s: value %v outside [%v, %v]", tc.name, got, tc.min, tc.max)
			}
		}
	}
}

func TestRandomCoordinateSinglePoint(t *testing.T) {
	if got := RandomCoordinate(10, 10); got != 10 {
		t.Errorf("got %v, want 10", got)
	}
}

func TestIsPublicIP(t *testing.T) {
	cases := []struct {
		ip       string
		expected bool
	}{
		{"8.8.8.8", true},
		{"2001:4860:4860::8888", true},
		{"192.168.1.1", false},
		{"10.0.0.1", false},
		{"172.16.5.4", false},
		{"127.0.0.1", false},
		{"::1", false},
		{"0.0.0.0", false},
		{"not-an-ip", false},
		{"", false},
	}

	for _, tc := range cases {
		if got := IsPublicIP(tc.ip); got != tc.expected {
			t.Errorf("%q: got %v, want %v", tc.ip, got, tc.expected)
		}
	}
}
