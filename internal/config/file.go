package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// DirName is the name of the grit configuration directory.
	DirName = ".grit"
	// ConfigFile is the name of the configuration file.
	ConfigFile = "config.yaml"
)

// Path returns the full path to the config file within the given root directory.
func Path(root string) string {
	return filepath.Join(root, DirName, ConfigFile)
}

// DirPath returns the full path to the .grit directory within the given root.
func DirPath(root string) string {
	return filepath.Join(root, DirName)
}

// Exists reports whether a grit configuration exists in the given root directory.
func Exists(root string) bool {
	_, err := os.Stat(Path(root))
	return err == nil
}

// Save writes the configuration to the .grit directory within the given root.
func Save(root string, cfg *Config) error {
	dirPath := DirPath(root)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(Path(root), data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// Load reads and parses the configuration from the given root directory.
func Load(root string) (*Config, error) {
	data, err := os.ReadFile(Path(root))
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// WriteGitignore creates a .gitignore file in the .grit directory.
func WriteGitignore(root string) error {
	gitignorePath := filepath.Join(DirPath(root), ".gitignore")
	content := "# Ignore credential files\n.credentials\n"
	return os.WriteFile(gitignorePath, []byte(content), 0644)
}

// FindRoot searches for a .grit directory starting from the current working
// directory and traversing up the directory tree.
func FindRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}

	for {
		if Exists(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a grit project (no .grit directory found)")
		}
		dir = parent
	}
}

// LoadFromWorkingDir finds the project root and loads its configuration.
func LoadFromWorkingDir() (*Config, error) {
	root, err := FindRoot()
	if err != nil {
		return nil, err
	}
	return Load(root)
}
