package sysinfo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// probeShellDashC probes the login shell: bash uses `-c 'echo $BASH_VERSION'`; other
// known POSIX shells use `{shell} --version` with the resolved binary and os.Environ().
func probeShellDashC(shellPath string) string {
	if shellPath == "" {
		return ""
	}

	base := strings.ToLower(strings.TrimSuffix(filepath.Base(shellPath), ".exe"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	switch base {
	case "bash":
		cmd := exec.CommandContext(ctx, shellPath, "-c", "echo $BASH_VERSION")
		cmd.Env = os.Environ()
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return ClipShellProbeOutput(string(out))
	}

	cmd := exec.CommandContext(ctx, shellPath, "--version")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return ClipShellProbeOutput(string(out))
}

func probePowerShellVersion(exe string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-NoProfile", "-Command", "$PSVersionTable.PSVersion.ToString()")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return ClipShellProbeOutput(string(out))
}
