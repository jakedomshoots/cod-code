//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type patchEnvelope struct {
	Patches []patchProposal `json:"patches"`
}

type patchProposal struct {
	Path string `json:"path"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

func main() {
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read prompt: %v\n", err)
		os.Exit(1)
	}
	text := string(prompt)
	switch os.Getenv("CEO_MODEL_REQUEST_KIND") {
	case "ceo_delegation":
		fmt.Println(`{"selected_subagents":["coder"],"summary":"Use the patch owner only for this narrow benchmark."}`)
		return
	case "ceo_review":
		if strings.Contains(text, "guard_verdict: pass") {
			fmt.Println(`{"recommended_verdict":"pass","summary":"Required checks passed after the model patch."}`)
			return
		}
		fmt.Println(`{"recommended_verdict":"fail","summary":"Guard verdict failed."}`)
		return
	}
	if !strings.Contains(text, "agent: coder") {
		fmt.Println("ok")
		return
	}
	oldText, err := os.ReadFile("internal/workspace/workspace.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read workspace fixture: %v\n", err)
		os.Exit(1)
	}
	envelope := patchEnvelope{
		Patches: []patchProposal{{
			Path: "internal/workspace/workspace.go",
			Old:  string(oldText),
			New:  pathEscapeFixedSource(),
		}},
	}
	if err := json.NewEncoder(os.Stdout).Encode(envelope); err != nil {
		fmt.Fprintf(os.Stderr, "write patch proposal: %v\n", err)
		os.Exit(1)
	}
}

func pathEscapeFixedSource() string {
	return `package workspace

import (
	"errors"
	"path/filepath"
	"strings"
)

var ErrPathEscapesWorkspace = errors.New("path escapes workspace")

func CleanRelativePath(path string) (string, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	if cleanPath == "." || cleanPath == ".." || filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return "", ErrPathEscapesWorkspace
	}
	return cleanPath, nil
}
`
}
