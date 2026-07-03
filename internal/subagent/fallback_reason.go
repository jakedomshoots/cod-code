package subagent

import (
	"errors"

	"ceoharness/internal/model"
)

func fallbackReasonForError(err error) string {
	var outputErr *OutputError
	if errors.As(err, &outputErr) && outputErr.Kind != "" {
		return string(outputErr.Kind)
	}
	var commandErr *model.CommandError
	if errors.As(err, &commandErr) {
		return string(commandErr.Kind)
	}
	var statusErr *model.HTTPStatusError
	if errors.As(err, &statusErr) {
		return string(statusErr.Kind)
	}
	return "provider_error"
}
