package models

// TopNInput contains a limit query param.
type TopNInput struct {
	Limit int `query:"limit" validate:"min=1,max=100"`
}
