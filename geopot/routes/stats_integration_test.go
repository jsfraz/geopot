package routes

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"jsfraz/geopot/database"
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
)

// TestAPIIntegration exercises the real router against a real database. It is
// skipped by default: `go test ./...` stays dependency free.
//
// Run it with the development environment loaded:
//
//	set -a && . ../.env.dev && set +a && GEOPOT_INTEGRATION=1 go test ./routes -run Integration -v
func TestAPIIntegration(t *testing.T) {
	if os.Getenv("GEOPOT_INTEGRATION") != "1" {
		t.Skip("set GEOPOT_INTEGRATION=1 with POSTGRES_*/VALKEY_* configured to run")
	}

	utils.LoadConfig()
	// Swagger is served from disk in debug mode, which the test directory does
	// not have.
	singleton := utils.GetSingleton()
	singleton.Config.GinMode = "release"

	database.SetupPostgres()
	database.SetupValkey()
	t.Cleanup(func() { _ = database.Close() })

	router, err := NewRouter()
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	// Query parameters are optional: an omitted limit/hours must fall back to
	// the documented default instead of returning 400.
	t.Run("defaults", func(t *testing.T) {
		cases := []struct {
			path  string
			check func(t *testing.T, body []byte)
		}{
			{
				path: "/api/stats/hourly",
				check: func(t *testing.T, body []byte) {
					var stats []models.HourlyStat
					if err := json.Unmarshal(body, &stats); err != nil {
						t.Fatalf("decoding hourly stats: %v (%s)", err, body)
					}
					if len(stats) != 24 {
						t.Errorf("got %d hourly buckets, want the 24h default", len(stats))
					}
				},
			},
			{
				path: "/api/stats/recentConnections",
				check: func(t *testing.T, body []byte) {
					var connections []models.Connection
					if err := json.Unmarshal(body, &connections); err != nil {
						t.Fatalf("decoding recent connections: %v (%s)", err, body)
					}
					if len(connections) > 100 {
						t.Errorf("got %d connections, want at most the 100 default", len(connections))
					}
				},
			},
			{
				path: "/api/stats/topCountries",
				check: func(t *testing.T, body []byte) {
					var entries []models.TopEntry
					if err := json.Unmarshal(body, &entries); err != nil {
						t.Fatalf("decoding top countries: %v (%s)", err, body)
					}
					if len(entries) > 15 {
						t.Errorf("got %d entries, want at most the 15 default", len(entries))
					}
				},
			},
			{
				path: "/api/stats/topUsernames",
				check: func(t *testing.T, body []byte) {
					var entries []models.TopEntry
					if err := json.Unmarshal(body, &entries); err != nil {
						t.Fatalf("decoding top usernames: %v (%s)", err, body)
					}
				},
			},
			{
				path: "/api/stats/selfInfo",
				check: func(t *testing.T, body []byte) {
					// null is a valid answer when the record is not known yet.
					if len(body) == 0 {
						t.Error("empty body")
					}
				},
			},
		}

		for _, tc := range cases {
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Errorf("%s: status %d, body %s", tc.path, recorder.Code, recorder.Body.String())
				continue
			}
			tc.check(t, recorder.Body.Bytes())
		}
	})

	// Out of range values are still rejected.
	t.Run("validation", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/stats/topCountries?limit=1000", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status %d, want 400 for an out of range limit", recorder.Code)
		}
	})

	t.Run("gzip", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/stats/allLatLng", nil)
		request.Header.Set("Accept-Encoding", "gzip")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status %d", recorder.Code)
		}
		if got := recorder.Header().Get("Content-Encoding"); got != "gzip" {
			t.Fatalf("Content-Encoding = %q, want gzip", got)
		}
		reader, err := gzip.NewReader(recorder.Body)
		if err != nil {
			t.Fatalf("invalid gzip response: %v", err)
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("decompressing: %v", err)
		}
		var latLngs []models.LatLng
		if err := json.Unmarshal(body, &latLngs); err != nil {
			t.Fatalf("decoding heatmap: %v (%s)", err, body)
		}
	})

	t.Run("hourly buckets", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/stats/hourly?hours=6", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		var stats []models.HourlyStat
		if err := json.Unmarshal(recorder.Body.Bytes(), &stats); err != nil {
			t.Fatalf("decoding: %v", err)
		}
		if len(stats) != 6 {
			t.Fatalf("got %d buckets, want 6", len(stats))
		}
		if got, want := stats[len(stats)-1].Bucket.UTC(), time.Now().UTC().Truncate(time.Hour); !got.Equal(want) {
			t.Errorf("last bucket = %s, want the current hour %s", got, want)
		}
	})
}
