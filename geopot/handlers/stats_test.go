package handlers

import (
	"testing"

	"jsfraz/geopot/models"

	"github.com/go-playground/validator/v10"
)

// The query inputs must accept a zero value (parameter omitted by the client)
// so the handler can fall back to its documented default, while still
// rejecting out-of-range values. Before this, `?limit=` omitted returned 400.
func TestTopNInputValidation(t *testing.T) {
	validate := validator.New()

	cases := []struct {
		limit   int
		wantErr bool
	}{
		{0, false},   // omitted -> default
		{1, false},   // lower bound
		{100, false}, // upper bound
		{-1, true},
		{101, true},
	}

	for _, tc := range cases {
		err := validate.Struct(&models.TopNInput{Limit: tc.limit})
		if (err != nil) != tc.wantErr {
			t.Errorf("limit %d: got error %v, want error %v", tc.limit, err, tc.wantErr)
		}
	}
}

func TestHourlyStatsInputValidation(t *testing.T) {
	validate := validator.New()

	cases := []struct {
		hours   int
		wantErr bool
	}{
		{0, false},   // omitted -> default
		{1, false},   // lower bound
		{720, false}, // upper bound
		{-1, true},
		{721, true},
	}

	for _, tc := range cases {
		err := validate.Struct(&models.HourlyStatsInput{Hours: tc.hours})
		if (err != nil) != tc.wantErr {
			t.Errorf("hours %d: got error %v, want error %v", tc.hours, err, tc.wantErr)
		}
	}
}

func TestRecentConnectionsInputValidation(t *testing.T) {
	validate := validator.New()

	cases := []struct {
		limit   int
		wantErr bool
	}{
		{0, false},   // omitted -> default
		{500, false}, // upper bound
		{501, true},
	}

	for _, tc := range cases {
		err := validate.Struct(&RecentConnectionsInput{Limit: tc.limit})
		if (err != nil) != tc.wantErr {
			t.Errorf("limit %d: got error %v, want error %v", tc.limit, err, tc.wantErr)
		}
	}
}
