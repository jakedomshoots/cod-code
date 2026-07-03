package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_ProviderProofScript_dryRunWritesKimiPlan(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--dry-run",
		"--provider", "kimi",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provider proof dry-run failed: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Provider Proof Evidence",
		"- Mode: dry-run",
		"- Provider: kimi",
		"scripts/kimi-model-command.sh",
		"| cross-language-js-state-reducer | planned |",
		"| cross-language-python-retry-policy | planned |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
	requireTextFile(t, filepath.Join(outputDir, "cross-language-js-state-reducer", "plan.md"))
	requireTextFile(t, filepath.Join(outputDir, "cross-language-python-retry-policy", "plan.md"))
}

func Test_ProviderProofScript_liveFailsWhenBenchmarkSummaryFails(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")
	binDir := t.TempDir()
	writeExecutableScript(t, filepath.Join(binDir, "kimi"), "#!/bin/sh\nexit 0\n")
	writeExecutableScript(t, filepath.Join(binDir, "go"), `#!/bin/sh
if [ "$1" = "build" ]; then
  exit 0
fi
if [ "$1" = "run" ]; then
  output=""
  task=""
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --output-dir)
        shift
        output="$1"
        ;;
      --local-agent-benchmark-task)
        shift
        task="$1"
        ;;
    esac
    shift || true
  done
  mkdir -p "$output"
  cat > "$output/summary.json" <<JSON
{"task_id":"$task","passed":0,"partial":0,"failed":1,"timed_out":0,"incomplete_evidence":1}
JSON
  exit 0
fi
exit 1
`)

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--provider", "kimi",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	cmd.Env = append(cmd.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("provider proof unexpectedly passed with failing benchmark summary:\n%s", string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	if !strings.Contains(index, "| cross-language-js-state-reducer | fail |") {
		t.Fatalf("index.md missing failing benchmark row:\n%s", index)
	}
}
