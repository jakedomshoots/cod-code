package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFakeAdapter(t *testing.T, dir string, name string, mode string) string {
	t.Helper()
	path := filepath.Join(dir, name+"-adapter")
	body := `#!/bin/sh
case "$CEO_HARNESS_ADAPTER_PROBE" in
version)
  echo "` + name + ` version 1.2.3"
  exit 0
  ;;
dry-run)
  case "` + mode + `" in
  valid)
    cat >/dev/null
    echo '{"status":"pass","summary":"` + name + ` patch ready","confidence":0.9,"evidence":["fake ` + name + `"],"patches":[{"path":"app.txt","old":"old","new":"new"}]}'
    exit 0
    ;;
  invalid)
    echo "adapter completed successfully"
    exit 0
    ;;
  hang)
    sleep 5
    exit 0
    ;;
  esac
  ;;
esac
echo "missing probe" >&2
exit 2
`
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake adapter: %v", err)
	}
	return path
}

func writeVersionTimeoutChildAdapter(t *testing.T, dir string, name string, childPIDPath string) string {
	t.Helper()
	path := filepath.Join(dir, name+"-adapter")
	statePath := filepath.Join(dir, name+"-version-state")
	body := `#!/bin/sh
case "$CEO_HARNESS_ADAPTER_PROBE" in
version)
  if [ ! -f "` + statePath + `" ]; then
    echo "first version probe timed out" > "` + statePath + `"
    sleep 5 &
    echo "$!" > "` + childPIDPath + `"
    wait
    exit 0
  fi
  echo "` + name + ` version 1.2.3"
  exit 0
  ;;
dry-run)
  cat >/dev/null
  echo '{"status":"pass","summary":"` + name + ` patch ready","confidence":0.9,"evidence":["fake ` + name + `"],"patches":[{"path":"app.txt","old":"old","new":"new"}]}'
  exit 0
  ;;
esac
echo "missing probe" >&2
exit 2
`
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write child timeout fake adapter: %v", err)
	}
	return path
}

func writeTransientVersionAdapter(t *testing.T, dir string, name string) string {
	t.Helper()
	path := filepath.Join(dir, name+"-adapter")
	statePath := filepath.Join(dir, name+"-version-state")
	body := `#!/bin/sh
case "$CEO_HARNESS_ADAPTER_PROBE" in
version)
  if [ ! -f "` + statePath + `" ]; then
    echo "first version probe timed out" > "` + statePath + `"
    sleep 5
    exit 0
  fi
  echo "` + name + ` version 1.2.3"
  exit 0
  ;;
dry-run)
  cat >/dev/null
  echo '{"status":"pass","summary":"` + name + ` patch ready","confidence":0.9,"evidence":["fake ` + name + `"],"patches":[{"path":"app.txt","old":"old","new":"new"}]}'
  exit 0
  ;;
esac
echo "missing probe" >&2
exit 2
`
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write transient fake adapter: %v", err)
	}
	return path
}
