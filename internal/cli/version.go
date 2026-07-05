package cli

import (
	"fmt"
	"io"
	"strings"
)

var (
	Version   = "dev"
	Commit    = ""
	BuildDate = ""
)

func runVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "cod %s\n", versionDetails())
	return err
}

func versionDetails() string {
	parts := []string{Version}
	if Commit != "" {
		parts = append(parts, "commit="+Commit)
	}
	if BuildDate != "" {
		parts = append(parts, "built="+BuildDate)
	}
	return strings.Join(parts, " ")
}
