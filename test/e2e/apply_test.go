//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// FIXME: When commit signing is enabled, tests may fail due to interactive input.
func TestApply(t *testing.T) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("failed to generate random id: %v", err)
	}
	id := hex.EncodeToString(b)

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("bulk", "apply", "--key", id, "--force", "./testdata/0001-update-timestamp.yml")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// TODO: Consider exposing PR ref directly from CLI output?
	branch := fmt.Sprintf("bulk/%s", id)
	n, err := getGitHubPullRequestNumberByBranch("loozhengyuan", "test-bulk", branch)
	if err != nil {
		t.Fatalf("failed to retrieve pr number: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	tkr := time.NewTicker(10 * time.Second)
	defer tkr.Stop()

	for {
		state, err := getGitHubPullRequestStateByID("loozhengyuan", "test-bulk", n)
		if err != nil {
			t.Fatalf("failed to retrieve pr state: %v", err)
		}

		if state == "MERGED" {
			break
		}
		if state != "OPEN" {
			t.Fatalf("unexpected pr state: %v", state)
		}

		select {
		case <-ctx.Done():
			t.Fatal("timeout waiting for pr to be merged")
			return
		case <-tkr.C:
		}
	}
}

func getGitHubPullRequestNumberByBranch(org, repo, branch string) (int, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("gh", "pr", "view", branch, "--repo", fmt.Sprintf("%s/%s", org, repo), "--json", "number", "--jq", ".number")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("run command: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	n, err := strconv.Atoi(strings.TrimSpace(stdout.String()))
	if err != nil {
		return 0, fmt.Errorf("parse output: %w", err)
	}
	return n, nil
}

func getGitHubPullRequestStateByID(org, repo string, id int) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(id), "--repo", fmt.Sprintf("%s/%s", org, repo), "--json", "state", "--jq", ".state")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run command: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}
