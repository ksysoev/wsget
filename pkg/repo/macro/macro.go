package macro

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/core/command"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string              `yaml:"version"`
	Source  string              `yaml:"source"`
	Macro   map[string][]string `yaml:"macro"`
	Domains []string            `yaml:"domains"`
}

type Macro struct {
	macro   map[string]*command.Templates
	domains []string
}

// NewMacro creates a new Macro instance with the specified domains.
// The domains parameter is a slice of strings representing the allowed domains for the macro.
// Returns a pointer to the newly created Macro instance.
func NewMacro(domains []string) *Macro {
	return &Macro{
		macro:   make(map[string]*command.Templates),
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
		return fmt.Errorf("duplicate macro: %s", name)
	}

	if len(rawCommands) == 0 {
		return fmt.Errorf("empty macro: %s", name)
	}

	templs, err := command.NewMacro(rawCommands)

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
			return fmt.Errorf("duplicate macro: %s", name)
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

	return nil, fmt.Errorf("unknown command: %s", name)
}

// GetNames returns a list of all macro names stored in the Macro instance.
// It does not take any parameters.
// It returns a slice of strings containing the names of the macros.
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

// LoadMacroForDomain loads and merges macros for a specific domain from YAML files in a given directory.
// It takes macroDir, a string specifying the directory path, and domain, a string specifying the target domain.
// It returns a pointer to a Macro containing merged macros for the domain, or an error in case of failure.
// Errors may occur if the directory cannot be read, files cannot be parsed, or macros fail to merge.
// Ignores non-YAML files, directories, and files without a matching domain.
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
				return nil, fmt.Errorf("fail to loading macro from file %s, %w ", file.Name(), err)
			}
		}
	}

	return macro, nil
}

func (m *Macro) Download(filepath, url string) error {
	resp, err := http.Get(url) //nolint:gosec // This is a CLI tool, and the URL is provided by the user

	if err != nil {
		return fmt.Errorf("fail to download macro: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to download macro: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fail to read macro file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("fail to unmarshal macro: %w", err)
	}

	if cfg.Version != "1" {
		return fmt.Errorf("unsupported macro version: %s", cfg.Version)
	}

	for name, rawCommands := range cfg.Macro {
		if err := m.AddCommands(name, rawCommands); err != nil {
			return err
		}
	}

	cfg.Source = url

	data, err = yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("fail to download macro: %w", err)
	}

	if !strings.HasSuffix(filepath, ".yaml") || !strings.HasSuffix(filepath, ".yml") {
		filepath += ".yml"
	}

	// Save the downloaded macro to the file
	if err := os.WriteFile(filepath, data, 0o600); err != nil {
		return fmt.Errorf("fail to download macro to file %s: %w", filepath, err)
	}

	return nil
}
