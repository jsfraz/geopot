package models

// TopEntry represents a single entry in a top-N ranking.
type TopEntry struct {
	Label      string  `json:"label"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}
