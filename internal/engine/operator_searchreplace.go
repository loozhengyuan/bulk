package engine

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// TODO: Should define its own struct?
func searchReplace(target []string, replacements []StepEditorReplacement) error {
	// TODO: Support excludes/negation?
	paths := make([]string, 0)
	for _, glob := range target {
		m, err := filepath.Glob(glob)
		if err != nil {
			return fmt.Errorf("glob match: %w", err)
		}
		paths = append(paths, m...)
	}

	for _, r := range replacements {
		re, err := regexp.Compile(r.Search)
		if err != nil {
			return fmt.Errorf("compile regex: %w", err)
		}

		for _, p := range paths {
			b, err := os.ReadFile(p)
			if err != nil {
				return nil
			}

			s := re.ReplaceAll(b, []byte(r.Replace))
			if bytes.Equal(b, s) {
				continue
			}

			// NOTE: Permissions are only used when creating file so it is
			// not used in this case because the file should already exist.
			if err := os.WriteFile(p, s, 0644); err != nil {
				return fmt.Errorf("write file: %w", err)
			}
		}
	}
	return nil
}
