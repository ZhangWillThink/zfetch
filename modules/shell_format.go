package modules

import (
	"path/filepath"
	"strings"
)

// ComposeShellDisplay returns a concise "shell + version" string for module output.
func ComposeShellDisplay(name, version string) string {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if version == "" {
		return name
	}
	v := sanitizeShellVersionOutput(name, version)
	if v == "" {
		return name
	}
	return strings.TrimSpace(name + " " + v)
}

func sanitizeShellVersionOutput(shellName, raw string) string {
	v := strings.TrimSpace(raw)
	sh := filepath.Base(strings.TrimSpace(shellName))
	if sh == "" {
		return v
	}

	// Drop a leading shell name as printed by `fish --version`, `bash --version`, etc.
	for {
		lv := strings.ToLower(v)
		ls := strings.ToLower(sh)
		if !strings.HasPrefix(lv, ls) {
			break
		}
		v = strings.TrimSpace(v[len(sh):])
		v = strings.TrimLeft(v, " ,:，.（）\t-+")
		if v == "" {
			return ""
		}
	}

	// Remove redundant "version / 版本" prefixes left on the line
	for {
		lv := strings.ToLower(v)
		if strings.HasPrefix(lv, "version") {
			v = strings.TrimSpace(v[len("version"):])
		} else if strings.HasPrefix(v, "版本") {
			v = strings.TrimSpace(strings.TrimPrefix(v, "版本"))
		} else {
			break
		}
		v = strings.TrimLeft(v, " :：，,\t")
	}

	return strings.TrimSpace(v)
}
