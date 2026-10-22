package models

// TopNInput contains a limit query param. An omitted limit is validated as
// empty and replaced by the handler's default.
type TopNInput struct {
	Limit int `query:"limit" validate:"omitempty,min=1,max=100"`
}
