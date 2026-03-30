package models

// HourlyStatsInput contains query parameters for the hourly stats endpoint.
type HourlyStatsInput struct {
	Hours int `query:"hours" validate:"min=1,max=720"`
}
