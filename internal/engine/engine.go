package engine

import (
	"fmt"
	"os"
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
