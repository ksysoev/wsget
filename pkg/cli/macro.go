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

func (m *Macro) merge(macro *Macro) error {
	for name, cmd := range macro.macro {
		if _, ok := m.macro[name]; ok {
			return fmt.Errorf("duplicate macro name: %s", name)
		}

		m.macro[name] = cmd
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

func LoadMacroForDomain(macroDir, domain string) (*Macro, error) {
	files, err := os.ReadDir(macroDir)
	if err != nil {
		log.Fatal(err)
	}

	var macro *Macro

	for _, file := range files {
		fileMacro, err := LoadMacro(macroDir + "/" + file.Name())

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
