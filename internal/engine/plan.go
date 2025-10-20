package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

type Plan struct {
	Version int    `json:"version"`
	ID      string `json:"id"`
	On      On     `json:"on"`
	Steps   []Step `json:"steps"`
	Commit  Commit `json:"commit"`
}

type On struct {
	Repositories []string `json:"repositories"`
}

type Step struct {
	ExecScript    *OperatorExecScript    `json:"script,omitempty"`
	SearchReplace *OperatorSearchReplace `json:"editor,omitempty"`
}

func (s *Step) GetOperator() (Operator, error) {
	// Ensure that only one operator is defined per step
	if s.ExecScript != nil && s.SearchReplace != nil {
		return nil, fmt.Errorf("multiple operators defined in a single step")
	}
	if s.ExecScript != nil {
		return s.ExecScript, nil
	}
	if s.SearchReplace != nil {
		return s.SearchReplace, nil
	}
	return nil, fmt.Errorf("unknown operator")
}

type StepEditorReplacement struct {
	Search  string `json:"search"`
	Replace string `json:"replace"`
}

type Commit struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (p *Plan) Inject(data TemplateContext) error {
	var err error
	if p.Commit.Title, err = data.RenderString(p.Commit.Title); err != nil {
		return fmt.Errorf("inject commit.title: %w", err)
	}
	if p.Commit.Body, err = data.RenderString(p.Commit.Body); err != nil {
		return fmt.Errorf("inject commit.body: %w", err)
	}
	if p.Steps != nil {
		for i, step := range p.Steps {
			if step.ExecScript != nil {
				if p.Steps[i].ExecScript.Run, err = data.RenderString(step.ExecScript.Run); err != nil {
					return fmt.Errorf("inject steps.%d.script.run: %w", i, err)
				}
			}
		}
	}
	return nil
}

func NewPlanFromJSON(r io.Reader) (*Plan, error) {
	var p Plan
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	return &p, nil
}

func NewPlanFromJSONFile(name string) (*Plan, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	return NewPlanFromJSON(f)
}

func NewPlanFromYAML(r io.Reader) (*Plan, error) {
	var p Plan
	if err := yaml.NewDecoder(r).Decode(&p); err != nil {
		return nil, fmt.Errorf("decode yaml: %w", err)
	}
	return &p, nil
}

func NewPlanFromYAMLFile(name string) (*Plan, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	return NewPlanFromYAML(f)
}
