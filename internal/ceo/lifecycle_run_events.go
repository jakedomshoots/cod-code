package ceo

type runEventLifecycleCursor struct {
	available map[LifecycleState]struct{}
	final     LifecycleState
}

func newRunEventLifecycleCursor(events []LifecycleEvent) runEventLifecycleCursor {
	available := map[LifecycleState]struct{}{}
	final := LifecycleCreated
	for _, event := range events {
		available[event.State] = struct{}{}
		final = event.State
	}
	return runEventLifecycleCursor{available: available, final: final}
}

func (c runEventLifecycleCursor) stateFor(state LifecycleState) LifecycleState {
	if _, ok := c.available[state]; ok {
		return state
	}
	return ""
}

func (c runEventLifecycleCursor) finalState() LifecycleState {
	if c.final == "" {
		return LifecycleCreated
	}
	return c.final
}
