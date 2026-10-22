package database

import (
	"os"
	"testing"
	"time"

	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
)

// TestStatsIntegration exercises the TimescaleDB specific queries (continuous
// aggregate refresh, hourly tail query, heatmap window) against a real
// database. It is skipped by default: `go test ./...` stays dependency free.
//
// Run it with the development environment loaded:
//
//	set -a && . ../.env.dev && set +a && GEOPOT_INTEGRATION=1 go test ./database -run Integration -v
func TestStatsIntegration(t *testing.T) {
	if os.Getenv("GEOPOT_INTEGRATION") != "1" {
		t.Skip("set GEOPOT_INTEGRATION=1 with POSTGRES_* configured to run")
	}

	utils.LoadConfig()
	SetupPostgres()
	t.Cleanup(func() { _ = Close() })

	// Must not block and must not fail on an already materialized aggregate.
	RefreshContinuousAggregates()

	stats, err := GetHourlyStats(24)
	if err != nil {
		t.Fatalf("GetHourlyStats: %v", err)
	}
	if len(stats) != 24 {
		t.Fatalf("got %d buckets, want 24 (one per hour, gaps filled)", len(stats))
	}
	for i, stat := range stats {
		if stat.Bucket.Minute() != 0 || stat.Bucket.Second() != 0 {
			t.Errorf("bucket %d is not hour aligned: %s", i, stat.Bucket)
		}
		if i > 0 && !stat.Bucket.Equal(stats[i-1].Bucket.Add(time.Hour)) {
			t.Errorf("bucket %d (%s) does not follow %s", i, stat.Bucket, stats[i-1].Bucket)
		}
	}
	// The newest bucket is the current, incomplete hour.
	if got, want := stats[len(stats)-1].Bucket, time.Now().UTC().Truncate(time.Hour); !got.Equal(want) {
		t.Errorf("last bucket = %s, want the current hour %s", got, want)
	}

	latLngs, err := GetAllLatLngs()
	if err != nil {
		t.Fatalf("GetAllLatLngs: %v", err)
	}
	var withCoordinates int64
	if err := utils.GetSingleton().Postgres.Model(&models.Connection{}).
		Where("latitude != 0 AND longitude != 0").
		Count(&withCoordinates).Error; err != nil {
		t.Fatalf("counting connections with coordinates: %v", err)
	}
	if withCoordinates > 0 && len(latLngs) == 0 {
		t.Error("heatmap is empty although connections with coordinates exist (aggregate was not backfilled)")
	}
	for _, latLng := range latLngs {
		if latLng.Intensity <= 0 {
			t.Errorf("non positive intensity: %v", latLng.Intensity)
		}
	}

	// The counters are cached, but the first read must still work.
	if _, err := GetTotalConnectionCount(); err != nil {
		t.Errorf("GetTotalConnectionCount: %v", err)
	}
	if _, err := GetUniqueIPCount(); err != nil {
		t.Errorf("GetUniqueIPCount: %v", err)
	}
	if _, err := GetUniqueCountryCount(); err != nil {
		t.Errorf("GetUniqueCountryCount: %v", err)
	}
	if _, err := GetLast24HourConnections(); err != nil {
		t.Errorf("GetLast24HourConnections: %v", err)
	}
}
