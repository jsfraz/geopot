package models

// Value represents a simple integer value, typically used for counts or totals.
type Value struct {
	Value int64 `json:"value" validate:"required"`
}
