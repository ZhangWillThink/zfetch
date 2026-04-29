//go:build linux

package sysinfo

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func GetOS() *OSInfo {
	info := &OSInfo{}
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		data, err = os.ReadFile("/usr/lib/os-release")
		if err != nil {
			info.Name = "Linux"
			return info
		}
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := strings.Trim(parts[1], `"`)
		switch key {
		case "NAME":
			info.Name = value
		case "VERSION":
			info.Version = value
		case "ID":
			info.ID = value
		}
	}

	if info.Name == "" {
		info.Name = "Linux"
	}
	return info
}

func GetKernel() *KernelInfo {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		return &KernelInfo{Name: "Linux"}
	}

	b2s := func(b [65]int8) string {
		n := 0
		for n < len(b) && b[n] != 0 {
			n++
		}
		bytes := make([]byte, n)
		for i := 0; i < n; i++ {
			bytes[i] = byte(b[i])
		}
		return string(bytes)
	}

	return &KernelInfo{
		Name:    b2s(uts.Sysname),
		Release: b2s(uts.Release),
		Version: b2s(uts.Version),
		Machine: b2s(uts.Machine),
	}
}

func GetCPU() *CPUInfo {
	info := &CPUInfo{}
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return info
	}

	cores := 0
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "model name":
			if info.Name == "" {
				info.Name = value
			}
		case "cpu cores":
			c, _ := strconv.Atoi(value)
			if c > cores {
				cores = c
			}
		case "processor":
			info.Cores++
		}
	}

	if cores > 0 {
		info.Cores = cores
	}
	if info.Name == "" {
		info.Name = "Unknown"
	}

	return info
}

func GetGPU() *GPUInfo {
	info := &GPUInfo{}

	drmDir := "/sys/class/drm"
	entries, err := os.ReadDir(drmDir)
	if err != nil {
		return info
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "card") {
			continue
		}

		vendorPath := drmDir + "/" + name + "/device/vendor"
		devicePath := drmDir + "/" + name + "/device/device"

		vendorData, _ := os.ReadFile(vendorPath)
		deviceData, _ := os.ReadFile(devicePath)

		vendor := strings.TrimSpace(string(vendorData))
		device := strings.TrimSpace(string(deviceData))

		if vendor != "" && device != "" {
			if info.Name != "" {
				info.Name += ", "
			}
			gpuName := detectGPUName(name)
			if gpuName != "" {
				info.Name += gpuName
			} else {
				info.Name += name
			}
		}
	}

	if info.Name == "" {
		out, err := exec.Command("lspci", "-mm").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if strings.Contains(line, "VGA") || strings.Contains(line, "3D") || strings.Contains(line, "Display") {
					parts := strings.Split(line, `"`)
					for i, p := range parts {
						if strings.TrimSpace(p) != "" && i+2 < len(parts) {
							info.Name = strings.TrimSpace(parts[i+1])
							break
						}
					}
					break
				}
			}
		}
	}

	if info.Name == "" {
		info.Name = "Unknown"
	}
	return info
}

func detectGPUName(card string) string {
	namePath := "/sys/class/drm/" + card + "/device/"
	entries, _ := os.ReadDir(namePath)
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "drm") {
			vendorPath := namePath + e.Name() + "/vendor_name"
			devicePath := namePath + e.Name() + "/device_name"
			vendorData, _ := os.ReadFile(vendorPath)
			deviceData, _ := os.ReadFile(devicePath)
			vendor := strings.TrimSpace(string(vendorData))
			dev := strings.TrimSpace(string(deviceData))
			if vendor != "" || dev != "" {
				return vendor + " " + dev
			}
		}
	}
	return ""
}

func GetMemory() *MemoryInfo {
	info := &MemoryInfo{}
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return info
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		val, _ := strconv.ParseUint(fields[1], 10, 64)
		valKB := val * 1024

		switch fields[0] {
		case "MemTotal:":
			info.Total = valKB
		case "MemAvailable:":
			info.Available = valKB
		}
	}

	if info.Total > 0 && info.Available > 0 {
		info.Used = info.Total - info.Available
	}
	return info
}

func GetDisk() *DiskInfo {
	info := &DiskInfo{Path: "/"}
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return info
	}

	info.Total = stat.Blocks * uint64(stat.Bsize)
	info.Available = stat.Bavail * uint64(stat.Bsize)
	info.Used = info.Total - info.Available
	return info
}

func GetUptime() *UptimeInfo {
	info := &UptimeInfo{}
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return info
	}

	parts := strings.Fields(string(data))
	if len(parts) == 0 {
		return info
	}

	up, _ := strconv.ParseFloat(parts[0], 64)
	info.Uptime = uint64(up)
	return info
}

func GetShell() *ShellInfo {
	info := &ShellInfo{}

	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/sh"
	}
	info.Path = shellPath

	parts := strings.Split(shellPath, "/")
	info.Name = parts[len(parts)-1]

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
		info.Version = strings.TrimSpace(string(out))
		if len(info.Version) > 30 {
			info.Version = info.Version[:30]
		}
	}

	return info
}

func GetPackages() *PackageInfo {
	info := &PackageInfo{}
	managers := map[string]string{
		"dpkg":    "/var/lib/dpkg/status",
		"rpm":     "/var/lib/rpm",
		"pacman":  "/var/lib/pacman/local",
		"flatpak": "/var/lib/flatpak/app",
		"snap":    "/var/lib/snapd/snaps",
		"nix":     "/nix/store",
	}

	for mgr, path := range managers {
		count := countPackages(path)
		if count > 0 {
			info.Managers = append(info.Managers, mgr)
			info.Count += count
		}
	}

	return info
}

func countPackages(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}

	count := 0
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if strings.HasSuffix(name, ".snap") {
			continue
		}
		count++
	}
	return count
}

func GetDE() *DEInfo {
	info := &DEInfo{}

	desktopEnv := []string{
		"XDG_CURRENT_DESKTOP",
		"DESKTOP_SESSION",
		"XDG_SESSION_DESKTOP",
	}

	for _, env := range desktopEnv {
		val := os.Getenv(env)
		if val != "" {
			info.Name = val
			break
		}
	}

	return info
}

func GetWM() *WMInfo {
	info := &WMInfo{}

	if val := os.Getenv("XDG_SESSION_TYPE"); val != "" {
		info.Name = val
	}

	return info
}

func GetTerminal() *TerminalInfo {
	info := &TerminalInfo{}

	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	if termProgram != "" {
		info.Name = termProgram
	} else {
		info.Name = term
	}

	if termProgramVersion := os.Getenv("TERM_PROGRAM_VERSION"); termProgramVersion != "" {
		info.Version = termProgramVersion
	}

	if info.Name == "" {
		info.Name = "unknown"
	}

	return info
}

func GetResolution() *ResolutionInfo {
	info := &ResolutionInfo{}

	displayVar := os.Getenv("DISPLAY")
	if displayVar == "" {
		return info
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "xrandr")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, " connected") {
				fields := strings.Fields(line)
				for i, f := range fields {
					if strings.Contains(f, "x") && strings.Contains(f, "+") && i+1 < len(fields) {
						res := fields[i] + " " + fields[i+1]
						info.Resolutions = append(info.Resolutions, res)
						break
					}
				}
			}
		}
	}

	return info
}
