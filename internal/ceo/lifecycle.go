package ceo

import (
	"errors"
	"fmt"
)

type LifecycleState string

const (
	LifecycleCreated        LifecycleState = "created"
	LifecyclePlanning       LifecycleState = "planning"
	LifecycleDelegated      LifecycleState = "delegated"
	LifecyclePatchPreviewed LifecycleState = "patch_previewed"
	LifecyclePatchApplied   LifecycleState = "patch_applied"
	LifecycleChecking       LifecycleState = "checking"
	LifecycleReviewing      LifecycleState = "reviewing"
	LifecycleNeedsInput     LifecycleState = "needs_input"
	LifecyclePassed         LifecycleState = "passed"
	LifecycleFailed         LifecycleState = "failed"
	LifecycleCanceled       LifecycleState = "canceled"
	LifecycleRecovered      LifecycleState = "recovered"
)

var ErrInvalidLifecycleTransition = errors.New("invalid lifecycle transition")

type LifecycleEvent struct {
	Index         int            `json:"index"`
	State         LifecycleState `json:"state"`
	PreviousState LifecycleState `json:"previous_state,omitempty"`
	Summary       string         `json:"summary,omitempty"`
}

type LifecycleMachine struct {
	state  LifecycleState
	events []LifecycleEvent
}

func NewLifecycleMachine() LifecycleMachine {
	return LifecycleMachine{}
}

func (m *LifecycleMachine) Transition(next LifecycleState, summary string) error {
	if next == "" {
		return fmt.Errorf("empty lifecycle state: %w", ErrInvalidLifecycleTransition)
	}
	if m.state == next {
		return nil
	}
	if !validLifecycleTransition(m.state, next) {
		return fmt.Errorf("%s -> %s: %w", m.state, next, ErrInvalidLifecycleTransition)
	}
	previous := m.state
	m.state = next
	m.events = append(m.events, LifecycleEvent{
		Index:         len(m.events) + 1,
		State:         next,
		PreviousState: previous,
		Summary:       summary,
	})
	return nil
}

func (m *LifecycleMachine) Cancel(reason error) error {
	summary := "job canceled"
	if reason != nil {
		summary = "job canceled: " + reason.Error()
	}
	return m.Transition(LifecycleCanceled, summary)
}

func (m LifecycleMachine) State() LifecycleState {
	return m.state
}

func (m LifecycleMachine) Events() []LifecycleEvent {
	return append([]LifecycleEvent(nil), m.events...)
}

type lifecycleInput struct {
	Recovered      bool
	Canceled       bool
	CanceledAt     LifecycleState
	CancelReason   string
	ResultStatuses []string
	CheckStatuses  []string
	PreviewCount   int
	AppliedCount   int
	Verdict        string
}

type lifecycleResult struct {
	State  LifecycleState
	Events []LifecycleEvent
}

func buildLifecycle(input lifecycleInput) lifecycleResult {
	machine := NewLifecycleMachine()
	applyLifecycleTransition(&machine, LifecycleCreated, "job created")
	if input.Recovered {
		applyLifecycleTransition(&machine, LifecycleRecovered, "job recovered from saved context")
	}
	applyLifecycleTransition(&machine, LifecyclePlanning, "job packet planned")
	if len(input.ResultStatuses) > 0 || input.CanceledAt == LifecycleDelegated {
		applyLifecycleTransition(&machine, LifecycleDelegated, "subagents delegated")
	}
	if input.Canceled {
		applyLifecycleTransition(&machine, LifecycleCanceled, lifecycleCanceledSummary(input.CancelReason))
		return lifecycleResult{State: machine.State(), Events: machine.Events()}
	}
	if hasStatus(input.ResultStatuses, "needs_input") || input.Verdict == "needs_input" {
		applyLifecycleTransition(&machine, LifecycleNeedsInput, "user input required")
		return lifecycleResult{State: machine.State(), Events: machine.Events()}
	}
	if input.PreviewCount > 0 {
		applyLifecycleTransition(&machine, LifecyclePatchPreviewed, "patch preview created")
	}
	if input.AppliedCount > 0 {
		applyLifecycleTransition(&machine, LifecyclePatchApplied, "patch applied")
	}
	if len(input.CheckStatuses) > 0 {
		applyLifecycleTransition(&machine, LifecycleChecking, "verification checks run")
	}
	if input.Verdict == "fail" && hasStatus(input.CheckStatuses, "fail") {
		applyLifecycleTransition(&machine, LifecycleFailed, "verification check failed")
		return lifecycleResult{State: machine.State(), Events: machine.Events()}
	}
	applyLifecycleTransition(&machine, LifecycleReviewing, "CEO final review")
	applyLifecycleTransition(&machine, lifecycleStateForVerdict(input.Verdict), "CEO final verdict "+input.Verdict)
	return lifecycleResult{State: machine.State(), Events: machine.Events()}
}

func lifecycleCanceledSummary(reason string) string {
	if reason == "" {
		return "job canceled"
	}
	return "job canceled: " + reason
}

func lifecycleStateForVerdict(verdict string) LifecycleState {
	switch verdict {
	case "pass":
		return LifecyclePassed
	case "fail":
		return LifecycleFailed
	case "needs_input":
		return LifecycleNeedsInput
	case "canceled":
		return LifecycleCanceled
	default:
		return LifecycleFailed
	}
}

func applyLifecycleTransition(machine *LifecycleMachine, next LifecycleState, summary string) {
	if err := machine.Transition(next, summary); err != nil {
		_ = machine.Transition(LifecycleFailed, "lifecycle transition failed: "+err.Error())
	}
}

func hasStatus(statuses []string, want string) bool {
	for _, status := range statuses {
		if status == want {
			return true
		}
	}
	return false
}

func validLifecycleTransition(previous LifecycleState, next LifecycleState) bool {
	switch previous {
	case "":
		return next == LifecycleCreated
	case LifecycleCreated:
		return next == LifecycleRecovered ||
			next == LifecyclePlanning ||
			next == LifecycleFailed ||
			next == LifecycleCanceled
	case LifecycleRecovered:
		return next == LifecyclePlanning || next == LifecycleFailed || next == LifecycleCanceled
	case LifecyclePlanning:
		return next == LifecycleDelegated ||
			next == LifecyclePatchPreviewed ||
			next == LifecyclePatchApplied ||
			next == LifecycleChecking ||
			next == LifecycleReviewing ||
			next == LifecycleFailed ||
			next == LifecycleCanceled
	case LifecycleDelegated:
		return next == LifecyclePatchPreviewed ||
			next == LifecyclePatchApplied ||
			next == LifecycleChecking ||
			next == LifecycleReviewing ||
			next == LifecycleNeedsInput ||
			next == LifecycleCanceled
	case LifecyclePatchPreviewed:
		return next == LifecyclePatchApplied ||
			next == LifecycleChecking ||
			next == LifecycleReviewing ||
			next == LifecyclePassed ||
			next == LifecycleFailed ||
			next == LifecycleCanceled
	case LifecyclePatchApplied:
		return next == LifecycleChecking ||
			next == LifecycleReviewing ||
			next == LifecyclePassed ||
			next == LifecycleFailed ||
			next == LifecycleCanceled
	case LifecycleChecking:
		return next == LifecycleReviewing || next == LifecycleFailed || next == LifecycleCanceled
	case LifecycleReviewing:
		return next == LifecyclePassed ||
			next == LifecycleFailed ||
			next == LifecycleNeedsInput ||
			next == LifecycleCanceled
	case LifecycleNeedsInput, LifecyclePassed, LifecycleFailed, LifecycleCanceled:
		return false
	default:
		return false
	}
}
