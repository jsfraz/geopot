package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// freeipapiPayload is a trimmed but structurally faithful response from
// https://docs.freeipapi.com/ - note "timeZones" being an array and "zipCode"
// being null.
const freeipapiPayload = `{
	"ipVersion": 4,
	"ipAddress": "8.8.8.8",
	"latitude": 37.422,
	"longitude": -122.085,
	"countryName": "United States",
	"countryCode": "US",
	"capital": "Washington D.C.",
	"phoneCodes": [1],
	"timeZones": ["America/Adak", "America/Los_Angeles"],
	"zipCode": null,
	"cityName": "Mountain View",
	"regionName": "California",
	"regionCode": "CA",
	"continent": "Americas",
	"continentCode": "AM",
	"currencies": ["USD"],
	"languages": ["en"],
	"asn": "15169",
	"asnOrganization": "Google LLC",
	"isProxy": false
}`

func newTestConnection() *Connection {
	return NewConnection("8.8.8.8", "root", "secret", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
}

func TestSetConnectionDetailsFreeIpApi(t *testing.T) {
	connection := newTestConnection()

	if err := connection.SetConnectionDetails(freeipapiPayload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if connection.IPVersion != 4 {
		t.Errorf("IPVersion = %d, want 4", connection.IPVersion)
	}
	if connection.Latitude != 37.422 || connection.Longitude != -122.085 {
		t.Errorf("coordinates = %v, %v", connection.Latitude, connection.Longitude)
	}
	if connection.CountryName != "United States" || connection.CountryCode != "US" {
		t.Errorf("country = %q, %q", connection.CountryName, connection.CountryCode)
	}
	// The API reports an array, the first entry is what we store.
	if connection.TimeZone != "America/Adak" {
		t.Errorf("TimeZone = %q, want America/Adak", connection.TimeZone)
	}
	if connection.ZipCode != "" {
		t.Errorf("ZipCode = %q, want empty for null", connection.ZipCode)
	}
	if connection.CityName != "Mountain View" || connection.RegionName != "California" {
		t.Errorf("city/region = %q, %q", connection.CityName, connection.RegionName)
	}
	if connection.Continent != "Americas" || connection.ContinentCode != "AM" {
		t.Errorf("continent = %q, %q", connection.Continent, connection.ContinentCode)
	}
	if connection.IsProxy {
		t.Error("IsProxy should be false")
	}
}

func TestSetConnectionDetailsAcceptsSingleTimeZone(t *testing.T) {
	connection := newTestConnection()
	payload := `{"timeZone": "+02:00", "zipCode": "11000"}`

	if err := connection.SetConnectionDetails(payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connection.TimeZone != "+02:00" {
		t.Errorf("TimeZone = %q, want +02:00", connection.TimeZone)
	}
	if connection.ZipCode != "11000" {
		t.Errorf("ZipCode = %q, want 11000", connection.ZipCode)
	}
}

func TestSetConnectionDetailsKeepsCredentials(t *testing.T) {
	connection := newTestConnection()
	if err := connection.SetConnectionDetails(freeipapiPayload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connection.User != "root" || connection.Password != "secret" {
		t.Errorf("credentials were overwritten: %q, %q", connection.User, connection.Password)
	}
}

func TestSetConnectionDetailsRejectsInvalidJSON(t *testing.T) {
	connection := newTestConnection()
	if err := connection.SetConnectionDetails("{"); err == nil {
		t.Error("expected an error for malformed JSON")
	}
}

func TestConnectionAttemptsAreNotSerializedWhenZero(t *testing.T) {
	connection := newTestConnection()
	encoded, err := json.Marshal(connection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(string(encoded), "attempts") {
		t.Errorf("zero attempts should be omitted: %s", encoded)
	}

	connection.Attempts = 2
	encoded, err = json.Marshal(connection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(encoded), `"attempts":2`) {
		t.Errorf("attempts should be serialized while queued: %s", encoded)
	}
}
