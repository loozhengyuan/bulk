package engine

import (
	"bytes"
	"fmt"
	"text/template"
)

type TemplateContext struct {
	Plan Plan
}

func (t *TemplateContext) RenderString(s string) (string, error) {
	tpl, err := template.New("field").Parse(s)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var b bytes.Buffer
	if err := tpl.Execute(&b, t); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return b.String(), nil
}
