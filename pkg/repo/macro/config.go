package macro

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type config struct {
	Version string              `yaml:"version"`
	Source  string              `yaml:"source"`
	Macro   map[string][]string `yaml:"macro"`
	Domains []string            `yaml:"domains"`
}

func newConfig(src io.Reader) (*config, error) {
	var cfg *config

	decoder := yaml.NewDecoder(src)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *config) SetSource(source string) {
	c.Source = source
}

func (c *config) CreateRepo() (*Repo, error) {
	repo := New(c.Domains)

	for name, rawCommands := range c.Macro {
		err := repo.AddCommands(name, rawCommands)
		if err != nil {
			return nil, fmt.Errorf("fail to add macro: %w", err)
		}
	}

	return repo, nil
}

func (c *config) validate() error {
	if c.Version != "1" {
		return fmt.Errorf("unsupported macro version: %s", c.Version)
	}

	if len(c.Domains) == 0 {
		return fmt.Errorf("domains are required")
	}

	if len(c.Macro) == 0 {
		return fmt.Errorf("macro commands are required")
	}

	return nil
}

func (c *config) WriteTo(w io.Writer) (err error) {
	enc := yaml.NewEncoder(w)
	defer func() {
		if e := enc.Close(); err == nil && e != nil {
			err = fmt.Errorf("fail to write config: %w", e)
		}
	}()

	return enc.Encode(c)
}
