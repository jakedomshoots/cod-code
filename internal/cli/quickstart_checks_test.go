package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func Test_quickstartCheckCommand_returns_python_pytest_when_pyproject_configures_pytest(t *testing.T) {
	// Given
	root := t.TempDir()
	pyproject := "[tool.pytest.ini_options]\npythonpath = [\".\"]\n"
	if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte(pyproject), 0o644); err != nil {
		t.Fatalf("write pyproject.toml: %v", err)
	}

	// When
	got, ok := quickstartCheckCommand(root)

	// Then
	if !ok {
		t.Fatal("quickstartCheckCommand returned no command, want pytest command")
	}
	want := []string{"python", "-m", "pytest"}
	if !slices.Equal(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}

func Test_quickstartCheckCommand_returns_uv_pytest_when_uv_lock_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	pyproject := "[tool.pytest.ini_options]\npythonpath = [\".\"]\n"
	if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte(pyproject), 0o644); err != nil {
		t.Fatalf("write pyproject.toml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "uv.lock"), nil, 0o644); err != nil {
		t.Fatalf("write uv.lock: %v", err)
	}

	// When
	got, ok := quickstartCheckCommand(root)

	// Then
	if !ok {
		t.Fatal("quickstartCheckCommand returned no command, want uv pytest command")
	}
	want := []string{"uv", "run", "pytest"}
	if !slices.Equal(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}

func Test_quickstartCheckCommand_returns_make_test_when_makefile_has_test_target(t *testing.T) {
	// Given
	root := t.TempDir()
	makefile := ".PHONY: test\n\ntest:\n\tgo test ./...\n"
	if err := os.WriteFile(filepath.Join(root, "Makefile"), []byte(makefile), 0o644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	// When
	got, ok := quickstartCheckCommand(root)

	// Then
	if !ok {
		t.Fatal("quickstartCheckCommand returned no command, want make test command")
	}
	want := []string{"make", "test"}
	if !slices.Equal(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}
