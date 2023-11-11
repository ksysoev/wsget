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

type Macro map[string]Executer

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

	macro := make(Macro)

	for name, rawCommands := range cfg.Macro {
		var commands []Executer

		for _, rawCommand := range rawCommands {
			cmd, err := CommandFactory(rawCommand, nil)
			if err != nil {
				return nil, err
			}

			commands = append(commands, cmd)
		}

		if len(commands) == 0 {
			return nil, fmt.Errorf("empty macro: %s", name)
		}

		macro[name] = commands[0]
	}

	return &macro, nil
}
