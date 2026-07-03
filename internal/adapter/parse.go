package adapter

import (
	"fmt"

	"ceoharness/internal/subagent"
)

type ParsedOutput struct {
	Summary    string
	PatchCount int
}

func ParseOutput(text string) (ParsedOutput, error) {
	output, err := subagent.ParseModelOutput(text)
	if err != nil {
		return ParsedOutput{}, &Error{Kind: ErrorKindInvalidOutput, Err: fmt.Errorf("parse structured output: %w", err)}
	}
	if !output.Structured {
		return ParsedOutput{}, &Error{Kind: ErrorKindInvalidOutput, Err: ErrInvalidOutput}
	}
	return ParsedOutput{
		Summary:    output.Summary,
		PatchCount: len(output.PatchProposals),
	}, nil
}
