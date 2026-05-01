package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/muhammadmuzzammil1998/jsonc"
)

type ModuleOptions map[string]interface{}

type Config struct {
	Structure  string                   `json:"structure"`
	Logo       string                   `json:"logo,omitempty"`
	Separator  string                   `json:"separator"`
	ColorKeys  string                   `json:"colorKeys"`
	ColorTitle string                   `json:"colorTitle"`
	Pipe       bool                     `json:"pipe"`
	Modules    map[string]ModuleOptions `json:"modules,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Structure:  "title:separator:os:kernel:uptime:packages:shell:resolution:de:wm:terminal:cpu:gpu:memory:swap:disk:host:battery:localip:locale",
		Separator:  ": ",
		ColorKeys:  "default",
		ColorTitle: "default",
	}
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := jsonc.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func FindPreset(name string) (string, error) {
	searchPaths := []string{
		name,
		filepath.Join(name),
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(homeDir, ".config", "zfetch", name+".jsonc"),
			filepath.Join(homeDir, ".config", "zfetch", name),
			filepath.Join(homeDir, ".local", "share", "zfetch", "presets", name+".jsonc"),
			filepath.Join(homeDir, ".local", "share", "zfetch", "presets", name),
		)
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("preset %q not found in search paths", name)
}

func FindDefaultConfig() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "zfetch", "config.jsonc")
}

func ListConfigPaths() []string {
	var paths []string
	homeDir, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths,
			filepath.Join(homeDir, ".config", "zfetch"),
			filepath.Join(homeDir, ".local", "share", "zfetch", "presets"),
		)
	}
	return paths
}
