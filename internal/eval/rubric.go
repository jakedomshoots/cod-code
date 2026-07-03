package eval

import (
	"fmt"
	"os"
	"strings"
)

var requiredRubricSections = []string{
	"artifact-first scoring",
	"self-report exclusion",
	"scoring dimensions",
	"verdicts",
	"evidence paths",
}

func LoadRubric(path string) (Rubric, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Rubric{}, fmt.Errorf("read rubric: %w", err)
	}
	lower := strings.ToLower(string(content))
	missing := make([]string, 0)
	for _, section := range requiredRubricSections {
		if !strings.Contains(lower, section) {
			missing = append(missing, section)
		}
	}
	if len(missing) > 0 {
		return Rubric{}, fmt.Errorf("%w: missing sections %s", ErrInvalidRubric, strings.Join(missing, ", "))
	}
	return Rubric{
		Path:             path,
		RequiredSections: append([]string(nil), requiredRubricSections...),
	}, nil
}
