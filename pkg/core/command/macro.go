package command

import (
	"bytes"
	"text/template"

	"github.com/ksysoev/wsget/pkg/core"
)

type Templates struct {
	list []*template.Template
}

// NewMacro creates a new Templates instance by parsing a list of string templates.
// It takes a parameter templates of type []string, representing raw string templates.
// It returns a pointer to a Templates instance populated with parsed templates.
// It returns an error if any of the provided templates fail to parse.
func NewMacro(rawTemplates []string) (*Templates, error) {
	tmpls := &Templates{}
	tmpls.list = make([]*template.Template, len(rawTemplates))

	for i, rawTempl := range rawTemplates {
		tmpl, err := template.New("macro").Parse(rawTempl)
		if err != nil {
			return nil, err
		}

		tmpls.list[i] = tmpl
	}

	return tmpls, nil
}

// GetExecuter generates an Executer based on the provided arguments and the templates in the Templates list.
// It takes args of type []string, representing input arguments for template execution.
// It returns a core.Executer initialized with the evaluated templates or an error if template execution fails.
// It returns an error if a template execution fails or if command creation from the template output fails.
// If a single template is evaluated, it returns the respective command; otherwise, returns a sequence of commands.
func (t *Templates) GetExecuter(args []string) (core.Executer, error) {
	data := struct {
		Args []string
	}{args}
	cmds := make([]core.Executer, len(t.list))

	for i, tmpl := range t.list {
		var output bytes.Buffer
		if err := tmpl.Execute(&output, data); err != nil {
			return nil, err
		}

		cmd, err := NewFactory(nil).Create(output.String())
		if err != nil {
			return nil, err
		}

		cmds[i] = cmd
	}

	if len(cmds) == 1 {
		return cmds[0], nil
	}

	return NewSequence(cmds), nil
}
