package engine

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var _ Operator = (*OperatorSearchReplace)(nil)

type OperatorSearchReplace struct {
	Target       []string                `json:"target"`
	Replacements []StepEditorReplacement `json:"replacements"`
}

func (op *OperatorSearchReplace) Validate() error {
	if len(op.Target) == 0 {
		return fmt.Errorf("target is not specified")
	}
	if len(op.Replacements) == 0 {
		return fmt.Errorf("replacements is not specified")
	}
	for i, r := range op.Replacements {
		if _, err := regexp.Compile(r.Search); err != nil {
			return fmt.Errorf("replacements.%d.search is not a valid regex: %w", i, err)
		}
	}
	return nil
}

func (op *OperatorSearchReplace) Apply(ctx OperatorContext) error {
	// TODO: Support excludes/negation?
	paths := make([]string, 0)
	for _, glob := range op.Target {
		m, err := filepath.Glob(filepath.Join(ctx.Dir, glob))
		if err != nil {
			return fmt.Errorf("glob match '%s': %w", glob, err)
		}
		paths = append(paths, m...)
	}

	for _, r := range op.Replacements {
		re, err := regexp.Compile(r.Search)
		if err != nil {
			return fmt.Errorf("compile regex: %w", err)
		}

		for _, p := range paths {
			b, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("read file %s: %w", p, err)
			}

			out := re.ReplaceAll(b, []byte(r.Replace))
			if bytes.Equal(b, out) {
				continue
			}

			// NOTE: Permissions are only used when creating file so it is
			// not used in this case because the file should already exist.
			if err := os.WriteFile(p, out, 0644); err != nil {
				return fmt.Errorf("write file: %w", err)
			}
		}
	}
	return nil
}
