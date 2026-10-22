package models

// HourlyStatsInput contains query parameters for the hourly stats endpoint.
// An omitted hours value is validated as empty and replaced by the handler's
// default.
type HourlyStatsInput struct {
	Hours int `query:"hours" validate:"omitempty,min=1,max=720"`
}
