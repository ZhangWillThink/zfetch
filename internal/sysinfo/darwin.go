//go:build darwin

package sysinfo

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func GetOS() *OSInfo {
	info := &OSInfo{Name: "macOS"}
	out, err := exec.Command("sw_vers", "-productName").Output()
	if err == nil {
		info.Name = strings.TrimSpace(string(out))
	}
	out, err = exec.Command("sw_vers", "-productVersion").Output()
	if err == nil {
		info.Version = strings.TrimSpace(string(out))
	}
	info.ID = "macos"
	return info
}

func GetKernel() *KernelInfo {
	info := &KernelInfo{Name: "Darwin"}
	out, err := exec.Command("uname", "-s").Output()
	if err == nil {
		info.Name = strings.TrimSpace(string(out))
	}
	out, err = exec.Command("uname", "-r").Output()
	if err == nil {
		info.Release = strings.TrimSpace(string(out))
	}
	out, err = exec.Command("uname", "-v").Output()
	if err == nil {
		info.Version = strings.TrimSpace(string(out))
	}
	out, err = exec.Command("uname", "-m").Output()
	if err == nil {
		info.Machine = strings.TrimSpace(string(out))
	}
	return info
}

func GetCPU() *CPUInfo {
	info := &CPUInfo{}
	out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
	if err == nil {
		info.Name = strings.TrimSpace(string(out))
	}
	out, err = exec.Command("sysctl", "-n", "hw.logicalcpu").Output()
	if err == nil {
		cores, _ := strconv.Atoi(strings.TrimSpace(string(out)))
		info.Cores = cores
	}
	if info.Name == "" {
		info.Name = "Unknown"
	}
	return info
}

func GetGPU() *GPUInfo {
	info := &GPUInfo{}
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Chipset Model:") {
				if info.Name != "" {
					info.Name += ", "
				}
				name := strings.TrimSpace(strings.TrimPrefix(line, "Chipset Model:"))
				info.Name += name
			}
		}
	}
	if info.Name == "" {
		info.Name = "Unknown"
	}
	return info
}

func GetMemory() *MemoryInfo {
	info := &MemoryInfo{}

	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err == nil {
		total, _ := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
		info.Total = total
	}

	out, err = exec.Command("vm_stat").Output()
	if err == nil {
		var pageSize uint64 = 16384
		var freePages, inactivePages, wiredPages, activePages uint64

		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "Mach Virtual Memory Statistics") {
				if idx := strings.Index(line, "page size of"); idx != -1 {
					rest := line[idx+len("page size of"):]
					rest = strings.TrimSpace(rest)
					rest = strings.TrimSuffix(rest, "bytes)")
					rest = strings.TrimSpace(rest)
					ps, _ := strconv.ParseUint(rest, 10, 64)
					if ps > 0 {
						pageSize = ps
					}
				}
				continue
			}

			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.TrimRight(value, ".")

			switch key {
			case "Pages free":
				freePages, _ = strconv.ParseUint(value, 10, 64)
			case "Pages inactive":
				inactivePages, _ = strconv.ParseUint(value, 10, 64)
			case "Pages wired down":
				wiredPages, _ = strconv.ParseUint(value, 10, 64)
			case "Pages active":
				activePages, _ = strconv.ParseUint(value, 10, 64)
			}
		}

		if info.Total > 0 {
			available := (freePages + inactivePages) * pageSize
			used := (activePages + wiredPages) * pageSize
			info.Used = used
			info.Available = available
		}
	}

	return info
}

func GetDisk() *DiskInfo {
	info := &DiskInfo{Path: "/"}

	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		out, err := exec.Command("df", "-k", "/").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				fields := strings.Fields(line)
				if len(fields) >= 4 && fields[len(fields)-1] == "/" {
					totalKB, _ := strconv.ParseUint(fields[1], 10, 64)
					availKB, _ := strconv.ParseUint(fields[3], 10, 64)
					info.Total = totalKB * 1024
					info.Available = availKB * 1024
					info.Used = info.Total - info.Available
					break
				}
			}
		}
		return info
	}

	info.Total = stat.Blocks * uint64(stat.Bsize)
	info.Available = stat.Bavail * uint64(stat.Bsize)
	info.Used = info.Total - info.Available
	return info
}

func GetUptime() *UptimeInfo {
	info := &UptimeInfo{}

	out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err == nil {
		line := strings.TrimSpace(string(out))
		parts := strings.Split(line, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "sec = ") {
				secStr := strings.TrimPrefix(part, "sec = ")
				bootSec, parseErr := strconv.ParseInt(secStr, 10, 64)
				if parseErr == nil {
					now := time.Now().Unix()
					info.Uptime = uint64(now - bootSec)
				}
				break
			}
		}
	}

	return info
}

func GetShell() *ShellInfo {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/zsh"
	}
	parts := strings.Split(shellPath, "/")
	info := &ShellInfo{
		Name: parts[len(parts)-1],
		Path: shellPath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var versionCmd string
	switch info.Name {
	case "bash":
		versionCmd = "echo $BASH_VERSION"
	case "zsh":
		versionCmd = "zsh --version"
	case "fish":
		versionCmd = "fish --version"
	default:
		versionCmd = info.Name + " --version"
	}

	cmd := exec.CommandContext(ctx, info.Name, "-c", versionCmd)
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(out))
		if len(version) > 30 {
			version = version[:30]
		}
		info.Version = version
	}

	return info
}

func GetPackages() *PackageInfo {
	info := &PackageInfo{}

	brewPrefixes := []string{"/opt/homebrew", "/usr/local"}
	for _, prefix := range brewPrefixes {
		cellarPath := prefix + "/Cellar"
		entries, err := os.ReadDir(cellarPath)
		if err == nil {
			count := 0
			for _, e := range entries {
				if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
					count++
				}
			}
			if count > 0 {
				info.Managers = append(info.Managers, "brew")
				info.Count += count
				break
			}
		}
	}

	macportsPath := "/opt/local/var/macports/registry/ports"
	entries, err := os.ReadDir(macportsPath)
	if err == nil {
		count := 0
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				count++
			}
		}
		if count > 0 {
			info.Managers = append(info.Managers, "macports")
			info.Count += count
		}
	}

	nixPath := "/nix/store"
	entries, err = os.ReadDir(nixPath)
	if err == nil {
		count := 0
		for _, e := range entries {
			name := e.Name()
			if !strings.HasPrefix(name, ".") {
				count++
			}
		}
		if count > 0 {
			info.Managers = append(info.Managers, "nix")
			info.Count += count
		}
	}

	return info
}

func GetDE() *DEInfo {
	info := &DEInfo{}

	desktopEnv := []string{
		"XDG_CURRENT_DESKTOP",
		"DESKTOP_SESSION",
	}

	for _, env := range desktopEnv {
		val := os.Getenv(env)
		if val != "" {
			info.Name = val
			break
		}
	}

	if info.Name == "" {
		info.Name = "Aqua"
	}

	return info
}

func GetWM() *WMInfo {
	info := &WMInfo{}

	if val := os.Getenv("XDG_SESSION_TYPE"); val != "" {
		info.Name = val
	}

	if info.Name == "" {
		info.Name = "Quartz Compositor"
	}

	return info
}

func GetTerminal() *TerminalInfo {
	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram == "" {
		termProgram = os.Getenv("TERM")
	}
	info := &TerminalInfo{Name: termProgram}

	if version := os.Getenv("TERM_PROGRAM_VERSION"); version != "" {
		info.Version = version
	}

	if info.Name == "" {
		info.Name = "unknown"
	}

	return info
}

func GetResolution() *ResolutionInfo {
	info := &ResolutionInfo{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "system_profiler", "SPDisplaysDataType")
	out, err := cmd.Output()
	if err != nil {
		return info
	}

	var currentDisplay string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Resolution:") {
			res := strings.TrimSpace(strings.TrimPrefix(line, "Resolution:"))
			info.Resolutions = append(info.Resolutions, res)
			currentDisplay = ""
		}

		if strings.Contains(strings.ToLower(line), "display") && strings.HasSuffix(line, ":") {
			currentDisplay = line
		}

		_ = currentDisplay
	}

	return info
}
