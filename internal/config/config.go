package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Preset struct {
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
}

type Config struct {
	DefaultFormat string            `json:"default_format"`
	OutputFile    string            `json:"output_file"`
	Presets       map[string]Preset `json:"presets"`
}

func DefaultConfig() Config {
	return Config{
		DefaultFormat: "xml",
		OutputFile:    "cxport_output",
		Presets:       make(map[string]Preset),
	}
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "cxport"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	if cfg.Presets == nil {
		cfg.Presets = make(map[string]Preset)
	}
	if cfg.DefaultFormat == "" {
		cfg.DefaultFormat = "xml"
	}
	if cfg.OutputFile == "" {
		cfg.OutputFile = "cxport_output"
	}

	return cfg, nil
}

func Save(cfg Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) SavePreset(name string, paths []string) {
	c.Presets[name] = Preset{Name: name, Paths: paths}
}

func (c *Config) DeletePreset(name string) {
	delete(c.Presets, name)
}

func (c *Config) GetPresetNames() []string {
	names := make([]string, 0, len(c.Presets))
	for name := range c.Presets {
		names = append(names, name)
	}
	return names
}
