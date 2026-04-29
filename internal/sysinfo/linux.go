//go:build linux

package sysinfo

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
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

func GetGPU() []*GPUInfo {
	var gpus []*GPUInfo

	drmDir := "/sys/class/drm"
	entries, err := os.ReadDir(drmDir)
	if err != nil {
		return gpus
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "card") {
			continue
		}

		cardPath := drmDir + "/" + name + "/device"
		vendorPath := cardPath + "/vendor"
		devicePath := cardPath + "/device"

		vendorData, _ := os.ReadFile(vendorPath)
		deviceData, _ := os.ReadFile(devicePath)

		if vendor := strings.TrimSpace(string(vendorData)); vendor == "" {
			continue
		}
		_ = deviceData

		gpuName := detectGPUName(name)
		if gpuName == "" {
			continue
		}

		gpu := &GPUInfo{Name: strings.TrimSpace(gpuName)}

		vramPath := cardPath + "/mem_info_vram_total"
		if vramData, err := os.ReadFile(vramPath); err == nil {
			bytes, _ := strconv.ParseUint(strings.TrimSpace(string(vramData)), 10, 64)
			gpu.MemoryMiB = int(bytes / (1024 * 1024))
		}

		bootVgaPath := cardPath + "/boot_vga"
		if bootData, err := os.ReadFile(bootVgaPath); err == nil {
			if strings.TrimSpace(string(bootData)) == "1" {
				gpu.Type = "Discrete"
			}
		}

		if gpu.Type == "" {
			driverPath := cardPath + "/driver"
			if resolved, err := filepath.EvalSymlinks(driverPath); err == nil {
				if strings.Contains(resolved, "i915") || strings.Contains(resolved, "amdgpu/apu") {
					gpu.Type = "Integrated"
				} else {
					gpu.Type = "Discrete"
				}
			}
		}

		gpus = append(gpus, gpu)
	}

	if len(gpus) == 0 {
		out, err := exec.Command("lspci", "-mm").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if strings.Contains(line, "VGA") || strings.Contains(line, "3D") || strings.Contains(line, "Display") {
					parts := strings.Split(line, `"`)
					if len(parts) >= 4 {
						gpus = append(gpus, &GPUInfo{Name: strings.TrimSpace(parts[3])})
					}
				}
			}
		}
	}

	return gpus
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

func GetDisk() []*DiskInfo {
	var disks []*DiskInfo

	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return disks
	}

	seen := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		device := fields[0]
		mountpoint := fields[1]
		fstype := fields[2]

		if !strings.HasPrefix(device, "/dev/") {
			continue
		}

		if seen[mountpoint] {
			continue
		}
		seen[mountpoint] = true

		isPseudo := false
		for _, prefix := range []string{"/dev/loop", "/dev/ram", "/dev/sr"} {
			if strings.HasPrefix(device, prefix) {
				isPseudo = true
				break
			}
		}
		if isPseudo {
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
			Filesystem: fstype,
		})
	}

	return disks
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

func GetHost() *HostInfo {
	info := &HostInfo{}

	productData, _ := os.ReadFile("/sys/class/dmi/id/product_name")
	if p := strings.TrimSpace(string(productData)); p != "" {
		info.Product = p
	}

	versionData, _ := os.ReadFile("/sys/class/dmi/id/product_version")
	if v := strings.TrimSpace(string(versionData)); v != "" {
		info.Version = v
	}

	return info
}

func GetSwap() *SwapInfo {
	info := &SwapInfo{}
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return info
	}

	var total, free uint64
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		valKB := val * 1024
		switch fields[0] {
		case "SwapTotal:":
			total = valKB
		case "SwapFree:":
			free = valKB
		}
	}

	info.Total = total
	if total > free {
		info.Used = total - free
	}
	return info
}

func GetBattery() *BatteryInfo {
	info := &BatteryInfo{}

	entries, err := os.ReadDir("/sys/class/power_supply")
	if err != nil {
		return info
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "BAT") && !strings.HasPrefix(name, "AC") && name != "ADP1" && name != "ACAD" {
			// check if it's a battery
			typePath := filepath.Join("/sys/class/power_supply", name, "type")
			typeData, _ := os.ReadFile(typePath)
			typeStr := strings.TrimSpace(string(typeData))
			if typeStr != "Battery" {
				continue
			}
		}

		capPath := filepath.Join("/sys/class/power_supply", name, "capacity")
		capData, err := os.ReadFile(capPath)
		if err != nil {
			continue
		}
		cap, _ := strconv.Atoi(strings.TrimSpace(string(capData)))
		info.Percentage = cap

		statusPath := filepath.Join("/sys/class/power_supply", name, "status")
		statusData, _ := os.ReadFile(statusPath)
		status := strings.TrimSpace(string(statusData))

		switch status {
		case "Charging":
			info.Status = "Charging"
		case "Discharging":
			info.Status = "Discharging"
		case "Full":
			info.Status = "Full"
		case "Not charging":
			info.Status = "AC Connected"
		default:
			info.Status = status
		}

		if info.Percentage > 0 {
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
