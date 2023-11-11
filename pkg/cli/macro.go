package cli

import (
	"fmt"
	"os"

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

func NewMacro(domains []string) *Macro {
	return &Macro{
		macro:   make(map[string]Executer),
		domains: domains,
	}
}

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

func (m *Macro) Get(name string) (Executer, error) {
	if cmd, ok := m.macro[name]; ok {
		return cmd, nil
	}

	return nil, fmt.Errorf("unknown command: %s", name)
}

func LoadMacro(path string) (*Macro, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Version != "1" {
		return nil, fmt.Errorf("unsupported macro version: %s", cfg.Version)
	}

	macroCfg := NewMacro(cfg.Domains)

	for name, rawCommands := range cfg.Macro {
		if err := macroCfg.AddCommands(name, rawCommands); err != nil {
			return nil, err
		}
	}

	return macroCfg, nil
}
