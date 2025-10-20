package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Repository struct {
	id     string // ID of the execution
	dir    string // Local worktree of the repository
	remote string // URL of the Git remote
	auto   bool   // Whether to skip confirmation prompts
}

func (r *Repository) ApplyAndPushChanges(title, body string, steps ...Step) error {
	exists, err := r.isRemoteBranchExists()
	if err != nil {
		return fmt.Errorf("check remote branch exists: %w", err)
	}
	if exists {
		return nil
	}

	// Set up working branch
	if _, err := r.Run("git", "fetch", "--depth", "1", "origin", "HEAD"); err != nil {
		return fmt.Errorf("clone repo: %w", err)
	}
	if _, err := r.Run("git", "switch", "--create", r.branch(), "FETCH_HEAD"); err != nil {
		return fmt.Errorf("checkout branch: %w", err)
	}

	// Apply changes
	for i, step := range steps {
		op, err := step.GetOperator()
		if err != nil {
			return fmt.Errorf("get operator for step %d: %w", i, err)
		}
		// NOTE: Target needs to be relative of working dir
		if err := op.Apply(OperatorContext{Dir: r.dir}); err != nil {
			return fmt.Errorf("apply operator for step %d: %w", i, err)
		}
	}

	// Stage changes and commit
	if _, err := r.Run("git", "add", "."); err != nil {
		return fmt.Errorf("add files: %w", err)
	}
	if _, err := r.Run("git", "commit", "--message", title, "--message", body, "--trailer", fmt.Sprintf("Idempotency-Key:%s", r.id)); err != nil {
		return fmt.Errorf("commit files: %w", err)
	}

	// Preview diffs and prompt confirmation
	// TODO: If supporting multi-commit, this will not work!
	if _, err := r.Run("git", "--no-pager", "show", "--stat", "--patch", "--pretty=fuller", "HEAD"); err != nil {
		return fmt.Errorf("preview commit: %w", err)
	}

	if !r.auto {
		confirm, err := promptConfirm("Would you like to proceed with the aforementioned changes?")
		if err != nil {
			return fmt.Errorf("prompt confirm: %w", err)
		}
		if !confirm {
			return fmt.Errorf("confirm: %v", confirm)
		}
	}

	// Push changes
	if _, err := r.Run("git", "push", "--set-upstream", "origin", r.branch()); err != nil {
		return fmt.Errorf("push files: %w", err)
	}
	return nil
}

func (r *Repository) CreateGitHubPullRequest(title, body string) error {
	exists, err := r.isGitHubPullRequestExists()
	if err != nil {
		return fmt.Errorf("check pr exists: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := r.Run("gh", "pr", "create", "--head", r.branch(), "--title", title, "--body", body, "--assignee", "@me"); err != nil {
		return fmt.Errorf("create pr: %w", err)
	}
	if _, err := r.Run("gh", "pr", "merge", "--auto", "--squash", "--delete-branch", r.branch()); err != nil {
		return fmt.Errorf("enable pr automerge: %w", err)
	}
	return nil
}

func (r *Repository) branch() string {
	return fmt.Sprintf("bulk/%s", r.id)
}

func (r *Repository) isRemoteBranchExists() (bool, error) {
	if _, err := r.Run("git", "ls-remote", "--exit-code", "--heads", "origin", r.branch()); err != nil {
		// NOTE: Exit code 2 means branch not found
		var e *exec.ExitError
		if errors.As(err, &e) && e.ExitCode() == 2 {
			return false, nil
		}
		return false, fmt.Errorf("list remote branch: %w", err)
	}
	return true, nil
}

func (r *Repository) isGitHubPullRequestExists() (bool, error) {
	o, err := r.Run("gh", "pr", "list", "--head", r.branch(), "--state", "open", "--json", "number", "--jq", "length")
	if err != nil {
		return false, fmt.Errorf("list github prs: %w", err)
	}

	c, err := strconv.Atoi(strings.TrimSpace(o))
	if err != nil {
		return false, fmt.Errorf("parse output: %w", err)
	}
	if c != 1 {
		return false, nil
	}
	return true, nil
}

func (r *Repository) RunContext(ctx context.Context, cmd string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer

	c := exec.CommandContext(ctx, cmd, args...)
	c.Dir = r.dir
	c.Stdout = &stdout
	c.Stderr = &stderr

	if err := c.Run(); err != nil {
		if o := strings.TrimSpace(stdout.String()); o != "" {
			fmt.Fprintln(os.Stderr, o)
		}
		if o := strings.TrimSpace(stderr.String()); o != "" {
			fmt.Fprintln(os.Stderr, o)
		}
		return stdout.String(), fmt.Errorf("run command: %w", err)
	}
	return stdout.String(), nil
}

func (r *Repository) Run(cmd string, args ...string) (string, error) {
	return r.RunContext(context.Background(), cmd, args...)
}

func (r *Repository) Close() error {
	if err := os.RemoveAll(r.dir); err != nil {
		return fmt.Errorf("remove working dir: %w", err)
	}
	return nil
}

func NewRepository(id, remote string, auto bool) (*Repository, error) {
	// TODO: Explore using cache dir with temp dir?
	d, err := os.MkdirTemp("", id) // TODO: Slugify remote for nicer name?
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	r := Repository{
		id:     id, // TODO: Derive the ID from input file hash?
		dir:    d,
		remote: remote,
		auto:   auto,
	}

	if _, err := r.Run("git", "init", "."); err != nil {
		return nil, fmt.Errorf("init repo: %w", err)
	}
	if _, err := r.Run("git", "remote", "add", "origin", remote); err != nil { // TODO: Cater to HTTPS or GitLab?
		return nil, fmt.Errorf("set remote: %w", err)
	}
	return &r, nil
}
