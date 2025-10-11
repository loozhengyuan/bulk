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
	Repositories      []string          `json:"repositories"`
	RepositoriesMatch RepositoriesMatch `json:"repositoriesMatch"`
}

type RepositoriesMatch struct {
	Search    string   `json:"search"`
	Extension string   `json:"extension"`
	Filename  string   `json:"filename"`
	Language  string   `json:"language"`
	Owners    []string `json:"owners"`
	Repos     []string `json:"repos"`
	Size      string   `json:"size"`
}

// TODO: Implement sum types?
type Step struct {
	Script StepScript `json:"script"`
	Editor StepEditor `json:"editor"`
}

type StepScript struct {
	Run string `json:"run"`
}

type StepEditor struct {
	Target       []string                `json:"target"`
	Replacements []StepEditorReplacement `json:"replacements"`
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
		for i := range p.Steps {
			if p.Steps[i].Script.Run, err = data.RenderString(p.Steps[i].Script.Run); err != nil {
				return fmt.Errorf("inject steps.%d.script.run: %w", i, err)
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
