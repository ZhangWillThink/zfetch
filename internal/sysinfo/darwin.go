//go:build darwin

package sysinfo

import (
	"context"
	"fmt"
	"net"
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

func GetGPU() []*GPUInfo {
	var gpus []*GPUInfo
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return gpus
	}

	lines := strings.Split(string(out), "\n")
	var currentGPU *GPUInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Chipset Model:") || strings.HasPrefix(line, "Model:") {
			if currentGPU != nil && currentGPU.Name != "" {
				gpus = append(gpus, currentGPU)
			}
			currentGPU = &GPUInfo{}
			if prefix := "Chipset Model:"; strings.HasPrefix(line, prefix) {
				currentGPU.Name = strings.TrimSpace(strings.TrimPrefix(line, prefix))
			} else {
				currentGPU.Name = strings.TrimSpace(strings.TrimPrefix(line, "Model:"))
			}
		}

		if currentGPU != nil {
			if strings.HasPrefix(line, "VRAM (Total):") || strings.HasPrefix(line, "VRAM (Dynamic, Max):") {
				vram := strings.TrimSpace(line[strings.Index(line, ":")+1:])
				vram = strings.TrimSuffix(vram, " MB")
				vram = strings.TrimSuffix(vram, " GB")
				if strings.Contains(line, "GB") {
					gb, _ := strconv.Atoi(vram)
					currentGPU.MemoryMiB = gb * 1024
				} else {
					mb, _ := strconv.Atoi(vram)
					currentGPU.MemoryMiB = mb
				}
			}

			if strings.HasPrefix(line, "Bus:") {
				bus := strings.TrimSpace(line[strings.Index(line, ":")+1:])
				if strings.Contains(bus, "Built-In") || strings.Contains(bus, "Integrated") {
					currentGPU.Type = "Integrated"
				} else {
					currentGPU.Type = "Discrete"
				}
			}

			if strings.HasPrefix(line, "Vendor:") {
				vendor := strings.TrimSpace(line[strings.Index(line, ":")+1:])
				if vendor == "Intel" || vendor == "Apple" {
					if currentGPU.Type == "" {
						currentGPU.Type = "Integrated"
					}
				}
				if vendor == "AMD" || vendor == "NVIDIA" || vendor == "Nvidia" {
					currentGPU.Type = "Discrete"
				}
			}
		}
	}

	if currentGPU != nil && currentGPU.Name != "" {
		gpus = append(gpus, currentGPU)
	}

	if len(gpus) == 0 {
		gpus = append(gpus, &GPUInfo{Name: "Unknown"})
	}
	return gpus
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

func GetDisk() []*DiskInfo {
	var disks []*DiskInfo

	out, err := exec.Command("df", "-k").Output()
	if err != nil {
		return disks
	}

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		device := fields[0]
		mountpoint := fields[len(fields)-1]

		if !strings.HasPrefix(device, "/dev/") {
			continue
		}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountpoint, &stat); err != nil {
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		available := stat.Bavail * uint64(stat.Bsize)
		if total == 0 {
			continue
		}

		disks = append(disks, &DiskInfo{
			Path:       mountpoint,
			Total:      total,
			Used:       total - available,
			Available:  available,
			Filesystem: fields[1],
		})
	}

	if len(disks) == 0 {
		var stat syscall.Statfs_t
		if err := syscall.Statfs("/", &stat); err == nil {
			total := stat.Blocks * uint64(stat.Bsize)
			available := stat.Bavail * uint64(stat.Bsize)
			disks = append(disks, &DiskInfo{
				Path:       "/",
				Total:      total,
				Used:       total - available,
				Available:  available,
				Filesystem: "apfs",
			})
		}
	}

	return disks
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

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Resolution:") {
			res := strings.TrimSpace(strings.TrimPrefix(line, "Resolution:"))
			info.Resolutions = append(info.Resolutions, res)
		}
	}

	return info
}

func GetHost() *HostInfo {
	info := &HostInfo{}

	out, err := exec.Command("sysctl", "-n", "hw.model").Output()
	if err == nil {
		info.Product = strings.TrimSpace(string(out))
	}

	return info
}

func GetSwap() *SwapInfo {
	info := &SwapInfo{}

	out, err := exec.Command("sysctl", "-n", "vm.swapusage").Output()
	if err == nil {
		line := strings.TrimSpace(string(out))
		parts := strings.Fields(line)
		for i, p := range parts {
			if p == "total" && i+2 < len(parts) {
				totalStr := strings.TrimSuffix(parts[i+2], "M")
				total, _ := strconv.ParseFloat(strings.TrimSuffix(totalStr, ".00"), 64)
				info.Total = uint64(total * 1024 * 1024)
			}
			if p == "used" && i+2 < len(parts) {
				usedStr := strings.TrimSuffix(parts[i+2], "M")
				used, _ := strconv.ParseFloat(strings.TrimSuffix(usedStr, ".00"), 64)
				info.Used = uint64(used * 1024 * 1024)
			}
		}
	}

	return info
}

func GetBattery() *BatteryInfo {
	info := &BatteryInfo{}

	out, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return info
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "%") {
			pctStr := ""
			for i, c := range line {
				if c == '%' {
					for j := i - 1; j >= 0 && line[j] >= '0' && line[j] <= '9'; j-- {
						pctStr = string(line[j]) + pctStr
					}
					break
				}
			}
			if pctStr != "" {
				info.Percentage, _ = strconv.Atoi(pctStr)
			}

			lower := strings.ToLower(line)
			if strings.Contains(lower, "charging") {
				info.Status = "Charging"
			} else if strings.Contains(lower, "discharging") {
				info.Status = "Discharging"
			} else if strings.Contains(lower, "charged") || strings.Contains(lower, "full") {
				info.Status = "Full"
			} else if strings.Contains(lower, "ac") {
				info.Status = "AC Connected"
			}
			break
		}
	}

	return info
}

func GetLocalIP() *LocalIPInfo {
	info := &LocalIPInfo{}

	ifaces, err := net.Interfaces()
	if err != nil {
		return info
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() || ipnet.IP.To4() == nil {
				continue
			}

			prefixLen, _ := ipnet.Mask.Size()
			entry := LocalIPEntry{
				Name: iface.Name,
				IP:   fmt.Sprintf("%s/%d", ipnet.IP.String(), prefixLen),
			}
			info.Interfaces = append(info.Interfaces, entry)
		}
	}

	return info
}

func GetLocale() *LocaleInfo {
	info := &LocaleInfo{}

	if lang := os.Getenv("LANG"); lang != "" {
		info.Locale = lang
	}

	return info
}
