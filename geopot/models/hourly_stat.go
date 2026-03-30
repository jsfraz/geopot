package models

import "time"

// HourlyStat represents connection count for a single hour bucket.
type HourlyStat struct {
	Bucket time.Time `json:"bucket"`
	Count  int64     `json:"count"`
}
