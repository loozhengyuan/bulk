package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	repos, err := e.getRepositories()
	if err != nil {
		return fmt.Errorf("get repositories: %w", err)
	}

	for _, repo := range repos {
		fmt.Printf("Processing %s...", repo)

		remote := fmt.Sprintf("git@github.com:%s.git", repo)
		r, err := NewRepository(e.p.ID, remote, e.force)
		if err != nil {
			return fmt.Errorf("new repo: %w", err)
		}
		defer r.Close() // TODO: Handle error?

		if err := r.ApplyAndPushChanges(e.p.Commit.Title, e.p.Commit.Body, e.p.Steps...); err != nil {
			return fmt.Errorf("apply and push changes: %w", err)
		}
		if err := r.CreateGitHubPullRequest(e.p.Commit.Title, e.p.Commit.Body); err != nil {
			return fmt.Errorf("create pr: %w", err)
		}
	}
	return nil
}

func (e *Engine) getRepositories() ([]string, error) {
	repoSet := make(map[string]struct{})
	for _, r := range e.p.On.Repositories {
		repoSet[r] = struct{}{}
	}

	if e.p.On.RepositoriesMatch.Search != "" {
		matchedRepos, err := e.searchRepositories()
		if err != nil {
			return nil, fmt.Errorf("search repositories: %w", err)
		}
		for _, r := range matchedRepos {
			repoSet[r] = struct{}{}
		}
	}

	if len(repoSet) == 0 {
		return nil, errors.New("no repos to process")
	}

	repos := make([]string, 0, len(repoSet))
	for r := range repoSet {
		repos = append(repos, r)
	}
	return repos, nil
}

func (e *Engine) searchRepositories() ([]string, error) {
	m := e.p.On.RepositoriesMatch
	args := []string{
		"search", "code", m.Search,
		"--json", "repository",
		"--jq", ".[].repository.nameWithOwner",
	}
	if m.Extension != "" {
		args = append(args, "--extension", m.Extension)
	}
	if m.Filename != "" {
		args = append(args, "--filename", m.Filename)
	}
	if m.Language != "" {
		args = append(args, "--language", m.Language)
	}
	for _, o := range m.Owners {
		args = append(args, "--owner", o)
	}
	for _, r := range m.Repos {
		args = append(args, "--repo", r)
	}
	if m.Size != "" {
		args = append(args, "--size", m.Size)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("gh", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if o := strings.TrimSpace(stdout.String()); o != "" {
			fmt.Fprintln(os.Stderr, o)
		}
		if o := strings.TrimSpace(stderr.String()); o != "" {
			fmt.Fprintln(os.Stderr, o)
		}
		return nil, fmt.Errorf("github search: %s", stderr.String())
	}
	return strings.Fields(stdout.String()), nil
}

func New(p *Plan) (*Engine, error) {
	c := TemplateContext{
		Plan: *p,
	}
	if err := p.Inject(c); err != nil {
		return nil, fmt.Errorf("inject tmpl: %w", err)
	}
	for i, step := range p.Steps {
		op, err := step.GetOperator()
		if err != nil {
			return nil, fmt.Errorf("get operator for step %d: %w", i, err)
		}
		if err := op.Validate(); err != nil {
			return nil, fmt.Errorf("validate step %d: %w", i, err)
		}
	}
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
