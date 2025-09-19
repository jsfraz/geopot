package models

// NumberValue represents a simple integer value, typically used for counts or totals.
type NumberValue struct {
	Value int64 `json:"value" validate:"required"`
}
