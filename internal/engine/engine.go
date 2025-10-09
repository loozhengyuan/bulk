package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type Engine struct {
	p     *Plan
	force bool
}

func (e *Engine) SetForce(force bool) {
	e.force = force
}

func (e *Engine) SetKey(key string) {
	// Override the plan ID if a non-empty key is provided
	k := strings.TrimSpace(key)
	if k != "" {
		e.p.ID = k
	}
}

func (e *Engine) Execute() error {
	// TODO: Implement logic to match repos
	for _, repo := range e.p.On.Repositories {
		if err := e.processRepo(repo); err != nil {
			return fmt.Errorf("process repo: %s: %w", repo, err)
		}
	}
	return nil
}

func (e *Engine) processRepo(repo string) error {
	// TODO: Derive the ID somewhere?
	// E.g. idempotency_key="$(printf '%s' "${script}" | sha256sum | awk '{ print $1 }')"
	branch := fmt.Sprintf("bulk/%s", e.p.ID)

	// Set up working dir
	// TODO: Perhaps create just 1 temp dir and share?
	d, err := os.MkdirTemp("", path.Base(repo)) // TODO: Improve file sanitisation
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(d)

	// Set up Git repo
	if err := execCommand(d, "git", "init", "."); err != nil {
		return fmt.Errorf("init repo: %w", err)
	}
	if err := execCommand(d, "git", "remote", "add", "origin", fmt.Sprintf("git@github.com:%s.git", repo)); err != nil { // TODO: Cater to HTTPS or GitLab?
		return fmt.Errorf("set remote: %w", err)
	}

	// Skip execution if branch already exists
	branchExist, err := e.checkBranchExist(d, branch)
	if err != nil {
		return fmt.Errorf("check remote branch exist: %w", err)
	}
	if branchExist {
		return fmt.Errorf("remote branch already exist: %s", branch)
	}

	// Clone repo
	if err := execCommand(d, "git", "fetch", "--depth", "1", "origin", "HEAD"); err != nil {
		return fmt.Errorf("clone repo: %w", err)
	}

	// Skip execution if PR already exists
	pullRequestExist, err := e.checkGitHubPullRequestExist(d, branch)
	if err != nil {
		return fmt.Errorf("check pull request exist: %w", err)
	}
	if pullRequestExist {
		return fmt.Errorf("pull request already exist: %s", branch)
	}

	// Switch to new branch
	if err := execCommand(d, "git", "switch", "--create", branch, "FETCH_HEAD"); err != nil {
		return fmt.Errorf("checkout branch: %w", err)
	}

	// Apply changes
	// TODO: Implement template vars?
	for _, step := range e.p.Steps {
		// TODO: Properly validate each step first to ensure
		// no conflicting properties?
		switch {
		case step.Script.Run != "":
			if err := execScript(d, step.Script.Run); err != nil {
				return fmt.Errorf("exec script: %w", err)
			}
		case len(step.Editor.Target) > 0 && len(step.Editor.Replacements) > 0:
			// NOTE: Target needs to be relative of working dir
			p := make([]string, 0, len(step.Editor.Target))
			for _, t := range step.Editor.Target {
				p = append(p, filepath.Join(d, t))
			}
			if err := searchReplace(p, step.Editor.Replacements); err != nil {
				return fmt.Errorf("search and replace: %w", err)
			}
		default:
			return fmt.Errorf("invalid step definition: %#v", step)
		}
	}

	// Stage changes and commit
	if err := execCommand(d, "git", "add", "."); err != nil {
		return fmt.Errorf("add files: %w", err)
	}
	if err := execCommand(d, "git", "commit", "--message", e.p.Commit.Title, "--message", e.p.Commit.Body, "--trailer", fmt.Sprintf("Idempotency-Key:%s", e.p.ID)); err != nil {
		return fmt.Errorf("commit files: %w", err)
	}

	// Preview diffs and prompt confirmation
	// TODO: If supporting multi-commit, this will not work!
	if err := execCommand(d, "git", "--no-pager", "show", "--stat", "--patch", "--pretty=fuller", "HEAD"); err != nil {
		return fmt.Errorf("preview commit: %w", err)
	}

	if !e.force {
		confirm, err := promptConfirm("Would you like to proceed with the aforementioned changes?")
		if err != nil {
			return fmt.Errorf("prompt confirm: %w", err)
		}
		if !confirm {
			return fmt.Errorf("confirm: %v", confirm)
		}
	}

	// Push changes
	if err := execCommand(d, "git", "push", "--set-upstream", "origin", branch); err != nil {
		return fmt.Errorf("push files: %w", err)
	}

	// Create PR
	// NOTE: Cannot use `--fill` due to "unknown revision or path"?
	if err := execCommand(d, "gh", "pr", "create", "--head", branch, "--title", e.p.Commit.Title, "--body", e.p.Commit.Body, "--assignee", "@me"); err != nil {
		return fmt.Errorf("create pr: %w", err)
	}
	if err := execCommand(d, "gh", "pr", "merge", "--auto", "--squash", "--delete-branch", branch); err != nil {
		return fmt.Errorf("enable pr automerge: %w", err)
	}
	return nil
}

func (e *Engine) checkBranchExist(dir string, branch string) (bool, error) {
	if err := execCommand(dir, "git", "ls-remote", "--exit-code", "--heads", "origin", branch); err != nil {
		// NOTE: Exit code 2 means branch not found
		var e *exec.ExitError
		if errors.As(err, &e) && e.ExitCode() == 2 {
			return false, nil
		}
		return false, fmt.Errorf("list remote branch: %w", err)
	}
	return true, nil
}

func (e *Engine) checkGitHubPullRequestExist(dir string, branch string) (bool, error) {
	var stdout, stderr bytes.Buffer

	c := exec.Command("gh", "pr", "list", "--head", branch, "--state", "open", "--json", "number", "--jq", "length")
	c.Dir = dir
	c.Stdout = &stdout
	c.Stderr = &stderr

	if err := c.Run(); err != nil {
		fmt.Fprint(os.Stdout, stdout.String())
		fmt.Fprint(os.Stderr, stderr.String())
		return false, fmt.Errorf("run command: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(stdout.String()))
	if err != nil {
		return false, fmt.Errorf("parse output: %w", err)
	}
	if count == 1 {
		return true, nil
	}
	return false, nil
}

func New(p *Plan) (*Engine, error) {
	return &Engine{p: p}, nil
}

func NewFromFile(name string) (*Engine, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	p, err := NewPlanFromYAML(f)
	if err != nil {
		return nil, fmt.Errorf("parse plan: %w", err)
	}
	return New(p)
}
