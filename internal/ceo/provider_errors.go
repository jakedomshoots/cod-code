package ceo

import (
	"errors"

	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
)

type providerErrorFields struct {
	kind         string
	httpStatus   int
	retryAfterMS int64
	modelSource  string
	providerName string
}

func providerErrorFieldsFrom(err error) providerErrorFields {
	fields := providerErrorFields{}
	var routeErr *subagent.RouteError
	if errors.As(err, &routeErr) {
		fields.modelSource = routeErr.Metadata.Source
		fields.providerName = routeErr.Metadata.ProviderName
	}
	var outputErr *subagent.OutputError
	if errors.As(err, &outputErr) && outputErr.Kind != "" {
		fields.kind = string(outputErr.Kind)
	}
	var statusErr *model.HTTPStatusError
	if errors.As(err, &statusErr) {
		fields.kind = string(statusErr.Kind)
		fields.httpStatus = statusErr.StatusCode
		fields.retryAfterMS = statusErr.RetryAfterMS
	}
	var commandErr *model.CommandError
	if errors.As(err, &commandErr) {
		fields.kind = string(commandErr.Kind)
	}
	return fields
}
