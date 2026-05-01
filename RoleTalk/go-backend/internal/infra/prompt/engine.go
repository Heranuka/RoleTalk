// Package prompt manages AI instructions using templates.
package prompt

import (
	"bytes"
	"fmt"
	"go-backend/internal/models/domain"
	"text/template"
)

// Engine handles the compilation and execution of AI prompt templates.
type Engine struct {
	templates *template.Template
}

// NewEngine pre-compiles all templates at startup for maximum performance.
func NewEngine() (*Engine, error) {
	tmpl := template.New("prompts")

	// Register all templates
	templates := map[string]string{
		"roleplay":   RoleplaySystemTemplate,
		"evaluation": EvaluationTemplate,
	}

	for name, content := range templates {
		if _, err := tmpl.New(name).Parse(content); err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}
	}

	return &Engine{templates: tmpl}, nil
}

// RenderRoleplay builds a system prompt for the AI dialog loop.
func (e *Engine) RenderRoleplay(params domain.RoleplayParams) (string, error) {
	return e.execute("roleplay", params)
}

// RenderEvaluation builds a prompt for the AI analyst after the session ends.
func (e *Engine) RenderEvaluation(params domain.EvaluationParams) (string, error) {
	return e.execute("evaluation", params)
}

// execute is an internal helper to run the template engine.
func (e *Engine) execute(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return buf.String(), nil
}
