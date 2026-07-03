package subagent

import "errors"

type OutputErrorKind string

const (
	OutputErrorKindEmpty   OutputErrorKind = "model_output_empty"
	OutputErrorKindInvalid OutputErrorKind = "model_output_invalid"
)

var errModelOutputEmpty = errors.New("model output is empty")

type OutputError struct {
	Kind OutputErrorKind
	Err  error
}

func newInvalidOutputError(err error) *OutputError {
	return &OutputError{Kind: OutputErrorKindInvalid, Err: err}
}

func newEmptyOutputError() *OutputError {
	return &OutputError{Kind: OutputErrorKindEmpty, Err: errModelOutputEmpty}
}

func (e *OutputError) Error() string {
	if e == nil {
		return string(OutputErrorKindInvalid)
	}
	if e.Err == nil {
		return string(effectiveOutputErrorKind(e.Kind))
	}
	if e.Kind == "" {
		return e.Err.Error()
	}
	return string(e.Kind) + ": " + e.Err.Error()
}

func (e *OutputError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func effectiveOutputErrorKind(kind OutputErrorKind) OutputErrorKind {
	if kind == "" {
		return OutputErrorKindInvalid
	}
	return kind
}
