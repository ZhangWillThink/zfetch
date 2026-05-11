package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/muhammadmuzzammil1998/jsonc"
)

type Config struct {
	Structure  string                 `json:"structure"`
	Logo       string                 `json:"logo,omitempty"`
	Separator  string                 `json:"separator"`
	ColorKeys  string                 `json:"colorKeys"`
	ColorTitle string                 `json:"colorTitle"`
	Pipe       bool                   `json:"pipe"`
	Modules    map[string]interface{} `json:"modules,omitempty"`
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
	var searchPaths []string

	if filepath.IsAbs(name) || (len(name) > 2 && name[0] == '.' && os.IsPathSeparator(name[1])) {
		searchPaths = append(searchPaths, name)
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

	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		searchPaths = append(searchPaths,
			filepath.Join(exeDir, "presets", name+".jsonc"),
			filepath.Join(exeDir, "presets", name),
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

// ListPresetNames returns preset names discoverable in the same places FindPreset searches
// (bundled presets next to the binary, then user config and share directories).
func ListPresetNames() []string {
	seen := make(map[string]bool)
	var names []string

	add := func(base string) {
		base = strings.TrimSpace(base)
		if base == "" {
			return
		}
		key := strings.ToLower(base)
		if seen[key] {
			return
		}
		seen[key] = true
		names = append(names, base)
	}

	scanDir := func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			n := e.Name()
			switch {
			case strings.HasSuffix(n, ".jsonc"):
				add(strings.TrimSuffix(n, ".jsonc"))
			case filepath.Ext(n) == "":
				add(n)
			}
		}
	}

	if exe, err := os.Executable(); err == nil {
		scanDir(filepath.Join(filepath.Dir(exe), "presets"))
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		scanDir(filepath.Join(homeDir, ".config", "zfetch"))
		scanDir(filepath.Join(homeDir, ".local", "share", "zfetch", "presets"))
	}

	sort.Strings(names)
	return names
}
