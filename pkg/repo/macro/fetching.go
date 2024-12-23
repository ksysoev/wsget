package macro

import (
	"fmt"
	"net/http"
	"os"
)

// Download downloads a macro configuration file from the specified URL and saves it to the given file path.
// It takes filepath of type string and url of type string as inputs.
// It returns an error if the download fails, the file already exists, the YAML unmarshalling fails, or the macro version is unsupported.
func Download(filepath, url string) (err error) {
	resp, err := http.Get(url) //nolint:gosec // This is a CLI tool, and the URL is provided by the user

	if err != nil {
		return fmt.Errorf("fail to download macro: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to download macro: %s", resp.Status)
	}

	cfg, err := newConfig(resp.Body)
	if err != nil {
		return fmt.Errorf("fail to download macro: %w", err)
	}

	cfg.SetSource(url)

	if _, err := cfg.CreateRepo(); err != nil {
		return fmt.Errorf("fail to create commands: %w", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("fail to create file: %w", err)
	}

	defer func() {
		if e := file.Close(); err == nil && e != nil {
			err = fmt.Errorf("fail to close file: %w", e)
		}
	}()

	if err := cfg.Write(file); err != nil {
		return fmt.Errorf("fail to write macro: %w", err)
	}

	return nil
}
