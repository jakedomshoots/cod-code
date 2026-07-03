package ceo

const DefaultMaxCEOIterations = 6

type ceoIterationBudget struct {
	max       int
	used      int
	exhausted bool
}

func newCEOIterationBudget(max int) *ceoIterationBudget {
	return &ceoIterationBudget{
		max:  normalizeMaxCEOIterations(max),
		used: 1,
	}
}

func normalizeMaxCEOIterations(value int) int {
	if value <= 0 {
		return DefaultMaxCEOIterations
	}
	return value
}

func consumeCEOIterationBudget(budget *ceoIterationBudget) bool {
	if budget == nil {
		return true
	}
	if budget.used >= budget.max {
		budget.exhausted = true
		return false
	}
	budget.used++
	return true
}
