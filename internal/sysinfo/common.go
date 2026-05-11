package sysinfo

import (
	"os"
	"strings"
)

// MaxShellVersionBytes caps raw version probes so module output stays compact on all platforms.
const MaxShellVersionBytes = 80

// ClipShellProbeOutput trims whitespace and clamps length consistently (bytes, ASCII-first).
func ClipShellProbeOutput(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= MaxShellVersionBytes {
		return s
	}
	return s[:MaxShellVersionBytes]
}

// IsExcludedFilesystem skips pseudo/in-memory overlays when listing mounts (POSIX).
func IsExcludedFilesystem(fstype string) bool {
	switch strings.ToLower(strings.TrimSpace(fstype)) {
	case "tmpfs", "devtmpfs", "devfs", "proc", "sysfs",
		"cgroup", "cgroup2", "devpts", "fusectl", "securityfs",
		"debugfs", "tracefs", "hugetlbfs", "mqueue", "overlay",
		"squashfs", "bpf", "configfs", "pstore", "efivarfs":
		return true
	default:
		return false
	}
}

// TerminalFromEnv reads the same VT environment variables everywhere (vscode, WT, Terminal.app, SSH TERM, etc.).
func TerminalFromEnv() (name string, version string) {
	tp := strings.TrimSpace(os.Getenv("TERM_PROGRAM"))
	t := strings.TrimSpace(os.Getenv("TERM"))
	if tp != "" {
		name = tp
	} else {
		name = t
	}
	version = strings.TrimSpace(os.Getenv("TERM_PROGRAM_VERSION"))
	return name, version
}

func terminalNameOrUnknown(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "unknown"
	}
	return name
}

// LocaleFromUnixEnv matches common Unix resolution order across Linux and Darwin.
func LocaleFromUnixEnv() string {
	for _, k := range []string{"LC_ALL", "LANG", "LC_MESSAGES"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}
