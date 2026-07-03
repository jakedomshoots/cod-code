package history

type LifecycleEvent struct {
	Index         int    `json:"index"`
	State         string `json:"state"`
	PreviousState string `json:"previous_state,omitempty"`
	Summary       string `json:"summary,omitempty"`
}
