package repo

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/core/command"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string              `yaml:"version"`
	Macro   map[string][]string `yaml:"macro"`
	Domains []string            `yaml:"domains"`
}

type Macro struct {
	macro   map[string]*MacroTemplates
	domains []string
}

// NewMacro creates a new Macro instance with the specified domains.
// The domains parameter is a slice of strings representing the allowed domains for the macro.
// Returns a pointer to the newly created Macro instance.
func NewMacro(domains []string) *Macro {
	return &Macro{
		macro:   make(map[string]*MacroTemplates),
		domains: domains,
	}
}

// AddCommands adds a new macro with the given name and commands to the Macro instance.
// If a macro with the same name already exists, it returns an error.
// If the rawCommands slice is empty, it returns an error.
// If the rawCommands slice has only one command, it adds the command directly to the macro.
// Otherwise, it creates a new Sequence with the commands and adds it to the macro.
func (m *Macro) AddCommands(name string, rawCommands []string) error {
	if _, ok := m.macro[name]; ok {
		return &command.ErrDuplicateMacro{name}
	}

	if len(rawCommands) == 0 {
		return command.ErrEmptyMacro{name}
	}

	templs, err := NewMacroTemplates(rawCommands)

	if err != nil {
		return err
	}

	m.macro[name] = templs

	return nil
}

// merge merges the given macro into the current macro.
// If a macro with the same name already exists, an error is returned.
func (m *Macro) merge(macro *Macro) error {
	for name, cmd := range macro.macro {
		if _, ok := m.macro[name]; ok {
			return &command.ErrDuplicateMacro{name}
		}

		m.macro[name] = cmd
	}

	return nil
}

// Get returns the Executer associated with the given name, or an error if the name is not found.
func (m *Macro) Get(name, argString string) (core.Executer, error) {
	if cmd, ok := m.macro[name]; ok {
		args := strings.Fields(argString)
		return cmd.GetExecuter(args)
	}

	return nil, &command.ErrUnknownCommand{name}
}

func (m *Macro) GetNames() []string {
	names := make([]string, 0, len(m.macro))

	for name := range m.macro {
		names = append(names, name)
	}

	return names
}

// LoadFromFile loads a macro configuration from a file at the given path.
// It returns a Macro instance and an error if the file cannot be read or parsed.
func LoadFromFile(path string) (*Macro, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Version != "1" {
		return nil, &command.ErrUnsupportedVersion{cfg.Version}
	}

	macroCfg := NewMacro(cfg.Domains)

	for name, rawCommands := range cfg.Macro {
		if err := macroCfg.AddCommands(name, rawCommands); err != nil {
			return nil, err
		}
	}

	return macroCfg, nil
}

// LoadMacroForDomain loads a macro for a given domain from a directory.
// It takes the directory path and the domain name as input parameters.
// It returns a pointer to a Macro struct and an error if any.
func LoadMacroForDomain(macroDir, domain string) (*Macro, error) {
	files, err := os.ReadDir(macroDir)
	if err != nil {
		log.Fatal(err)
	}

	var macro *Macro

	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		fileMacro, err := LoadFromFile(macroDir + "/" + file.Name())

		if err != nil {
			return nil, err
		}

		hasDomain := false

		for _, fileDomain := range fileMacro.domains {
			if strings.HasSuffix(domain, fileDomain) {
				hasDomain = true
				break
			}
		}

		if !hasDomain {
			continue
		}

		if macro == nil {
			macro = fileMacro
		} else {
			err := macro.merge(fileMacro)

			if err != nil {
				return nil, fmt.Errorf("fail to loading macro from file %s, %s ", file.Name(), err)
			}
		}
	}

	return macro, nil
}

type MacroTemplates struct {
	list []*template.Template
}

func NewMacroTemplates(templates []string) (*MacroTemplates, error) {
	tmpls := &MacroTemplates{}
	tmpls.list = make([]*template.Template, len(templates))

	for i, rawTempl := range templates {
		tmpl, err := template.New("macro").Parse(rawTempl)
		if err != nil {
			return nil, err
		}

		tmpls.list[i] = tmpl
	}

	return tmpls, nil
}

func (t *MacroTemplates) GetExecuter(args []string) (core.Executer, error) {
	data := struct {
		Args []string
	}{args}
	cmds := make([]core.Executer, len(t.list))

	for i, tmpl := range t.list {
		var output bytes.Buffer
		if err := tmpl.Execute(&output, data); err != nil {
			return nil, err
		}

		cmd, err := command.NewFactory(nil).Create(output.String())
		if err != nil {
			return nil, err
		}

		cmds[i] = cmd
	}

	if len(cmds) == 1 {
		return cmds[0], nil
	}

	return command.NewSequence(cmds), nil
}
