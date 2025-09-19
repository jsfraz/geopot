package models

// StringsValue represents a simple string value, typically used for IPs.
type StringsValue struct {
	Value []string `json:"value" validate:"required"`
}
