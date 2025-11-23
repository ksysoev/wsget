package macro

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// config represents the configuration structure used for YAML parsing and validation.
// It contains fields for the version, source file, macros, and associated domains.
type config struct {
	Version string              `yaml:"version"`
	Source  string              `yaml:"source,omitempty"`
	Macro   map[string][]string `yaml:"macro"`
	Domains []string            `yaml:"domains"`
}

// newConfig creates and initializes a new config object from the provided YAML input.
// It takes src of type io.Reader which contains the YAML configuration data.
// It returns a pointer to a config instance and an error if the decoding or validation of the configuration fails.
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

// SetSource sets the Source field of the config struct to the provided string value.
// It takes source of type string as input and updates the Source field of the receiver.
// It does not return any values and does not perform validation on the input.
func (c *config) SetSource(source string) {
	c.Source = source
}

// CreateRepo initializes and returns a new Repo based on the config's domains and macros.
// It returns a pointer to a Repo instance and an error.
// It returns an error if adding any macro commands to the Repo fails.
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

// validate ensures that the config structure is properly initialized and contains valid data.
// It returns an error if the Version is unsupported, Domains are empty, or Macro commands are missing.
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

// Write encodes the config structure in YAML format and writes it to the provided io.Writer.
// It takes w of type io.Writer as input.
// It returns an error if the YAML encoding fails or if closing the encoder encounters an error.
func (c *config) Write(w io.Writer) (err error) {
	enc := yaml.NewEncoder(w)

	defer func() {
		if e := enc.Close(); err == nil && e != nil {
			err = fmt.Errorf("fail to write config: %w", e)
		}
	}()

	return enc.Encode(c)
}
