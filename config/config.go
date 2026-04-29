package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func parseJSONC(data []byte) ([]byte, error) {
	var result strings.Builder
	result.Grow(len(data))

	inString := false
	escaping := false
	inSingleComment := false
	inMultiComment := false

	for i := 0; i < len(data); i++ {
		b := data[i]

		if inSingleComment {
			if b == '\n' {
				inSingleComment = false
				result.WriteByte(b)
			}
			continue
		}

		if inMultiComment {
			if b == '*' && i+1 < len(data) && data[i+1] == '/' {
				inMultiComment = false
				i++
			}
			continue
		}

		if inString {
			if escaping {
				escaping = false
			} else if b == '\\' {
				escaping = true
			} else if b == '"' {
				inString = false
			}
			result.WriteByte(b)
			continue
		}

		if b == '"' {
			inString = true
			result.WriteByte(b)
			continue
		}

		if b == '/' && i+1 < len(data) {
			next := data[i+1]
			if next == '/' {
				inSingleComment = true
				i++
				continue
			}
			if next == '*' {
				inMultiComment = true
				i++
				continue
			}
		}

		result.WriteByte(b)
	}

	return []byte(result.String()), nil
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cleanData, err := parseJSONC(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSONC: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(cleanData, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
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
