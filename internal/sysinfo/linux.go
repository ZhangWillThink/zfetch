//go:build linux

package sysinfo

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
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
	h, err := host.Info()
	if err != nil {
		return &KernelInfo{Name: "Linux"}
	}
	return &KernelInfo{
		Name:    "Linux",
		Release: h.KernelVersion,
		Machine: h.KernelArch,
	}
}

func GetCPU() *CPUInfo {
	cpus, err := cpu.Info()
	if err != nil || len(cpus) == 0 {
		return &CPUInfo{Name: "Unknown"}
	}
	physicalCores, _ := cpu.Counts(false)
	info := &CPUInfo{
		Name:   cpus[0].ModelName,
		MaxMHz: cpus[0].Mhz,
	}
	if physicalCores > 0 {
		info.Cores = physicalCores
	} else {
		info.Cores = len(cpus)
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
	v, err := mem.VirtualMemory()
	if err != nil {
		return &MemoryInfo{}
	}
	return &MemoryInfo{
		Total:     v.Total,
		Used:      v.Used,
		Available: v.Available,
	}
}

func GetDisk() []*DiskInfo {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	var disks []*DiskInfo
	seen := map[string]bool{}

	for _, p := range partitions {
		if seen[p.Mountpoint] {
			continue
		}
		seen[p.Mountpoint] = true

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		disks = append(disks, &DiskInfo{
			Path:       p.Mountpoint,
			Total:      usage.Total,
			Used:       usage.Used,
			Available:  usage.Free,
			Filesystem: p.Fstype,
		})
	}

	return disks
}

func GetUptime() *UptimeInfo {
	u, err := host.Uptime()
	if err != nil {
		return &UptimeInfo{}
	}
	return &UptimeInfo{Uptime: u}
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
	s, err := mem.SwapMemory()
	if err != nil {
		return &SwapInfo{}
	}
	return &SwapInfo{
		Total: s.Total,
		Used:  s.Used,
	}
}

func GetBattery() *BatteryInfo {
	batteries, err := battery.GetAll()
	if err != nil || len(batteries) == 0 {
		return &BatteryInfo{}
	}
	b := batteries[0]
	info := &BatteryInfo{}
	if b.Full > 0 {
		info.Percentage = int(b.Current / b.Full * 100)
	}
	switch b.State.Raw {
	case battery.Charging:
		info.Status = "Charging"
	case battery.Discharging:
		info.Status = "Discharging"
	case battery.Full:
		info.Status = "Full"
	case battery.Idle:
		info.Status = "AC Connected"
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
