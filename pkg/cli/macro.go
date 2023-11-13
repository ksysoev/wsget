package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string              `yaml:"version"`
	Macro   map[string][]string `yaml:"macro"`
	Domains []string            `yaml:"domains"`
}

type Macro struct {
	macro   map[string]Executer
	domains []string
}

// NewMacro creates a new Macro instance with the specified domains.
// The domains parameter is a slice of strings representing the allowed domains for the macro.
// Returns a pointer to the newly created Macro instance.
func NewMacro(domains []string) *Macro {
	return &Macro{
		macro:   make(map[string]Executer),
		domains: domains,
	}
}

// AddCommands adds a new macro with the given name and commands to the Macro instance.
// If a macro with the same name already exists, it returns an error.
// If the rawCommands slice is empty, it returns an error.
// If the rawCommands slice has only one command, it adds the command directly to the macro.
// Otherwise, it creates a new CommandSequence with the commands and adds it to the macro.
func (m *Macro) AddCommands(name string, rawCommands []string) error {
	if _, ok := m.macro[name]; ok {
		return fmt.Errorf("macro already exists: %s", name)
	}

	commands := []Executer{}

	for _, rawCommand := range rawCommands {
		cmd, err := CommandFactory(rawCommand, nil)
		if err != nil {
			return err
		}

		commands = append(commands, cmd)
	}

	switch len(commands) {
	case 0:
		return fmt.Errorf("empty macro: %s", name)
	case 1:
		m.macro[name] = commands[0]
	default:
		m.macro[name] = NewCommandSequence(commands)
	}

	return nil
}

// merge merges the given macro into the current macro.
// If a macro with the same name already exists, an error is returned.
func (m *Macro) merge(macro *Macro) error {
	for name, cmd := range macro.macro {
		if _, ok := m.macro[name]; ok {
			return fmt.Errorf("duplicate macro name: %s", name)
		}

		m.macro[name] = cmd
	}

	return nil
}

// Get returns the Executer associated with the given name, or an error if the name is not found.
func (m *Macro) Get(name string) (Executer, error) {
	if cmd, ok := m.macro[name]; ok {
		return cmd, nil
	}

	return nil, fmt.Errorf("unknown command: %s", name)
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
		return nil, fmt.Errorf("unsupported macro file version: %s", path)
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
