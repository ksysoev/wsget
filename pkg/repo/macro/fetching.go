package macro

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Download downloads a macro configuration file from the specified URL and saves it to the given file path.
// It takes filepath of type string and url of type string as inputs.
// It returns an error if the download fails, the file already exists, the YAML unmarshalling fails, or the macro version is unsupported.
func Download(filepath, url string) error {
	resp, err := http.Get(url) //nolint:gosec // This is a CLI tool, and the URL is provided by the user

	if err != nil {
		return fmt.Errorf("fail to download macro: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to download macro: %s", resp.Status)
	}

	_, cfg, err := parseConfig(resp.Body)
	if err != nil {
		return fmt.Errorf("fail to download macro: %w", err)
	}

	cfg.Source = url

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("fail to download macro: %w", err)
	}

	if !strings.HasSuffix(filepath, ".yaml") || !strings.HasSuffix(filepath, ".yml") {
		filepath += ".yml"
	}

	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("file %s already exists, please use update command or use different name", filepath)
	}

	// Save the downloaded macro to the file
	if err := os.WriteFile(filepath, data, 0o600); err != nil {
		return fmt.Errorf("fail to download macro to file %s: %w", filepath, err)
	}

	return nil
}
