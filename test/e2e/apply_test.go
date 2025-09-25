//go:build e2e

package e2e

import (
	"bytes"
	"os/exec"
	"testing"
)

// FIXME: When commit signing is enabled, tests may fail due to interactive input.
func TestApply(t *testing.T) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("bulk", "apply", "--force", "./testdata/0001-bump-version.yml")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// TODO: Assert PR and commit in `main`
}
