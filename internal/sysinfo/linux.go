//go:build linux

package sysinfo

import (
	"context"
	"os"
	"os/exec"
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

func isPseudoFs(fstype string) bool {
	switch fstype {
	case "tmpfs", "devtmpfs", "devfs", "proc", "sysfs",
		"cgroup", "cgroup2", "devpts", "fusectl", "securityfs",
		"debugfs", "tracefs", "hugetlbfs", "mqueue", "overlay",
		"squashfs", "bpf", "configfs", "pstore", "efivarfs":
		return true
	}
	return false
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

		if isPseudoFs(p.Fstype) {
			continue
		}

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
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
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

	if c := countDpkg(); c > 0 {
		info.Managers = append(info.Managers, "dpkg")
		info.Count += c
	}
	if c := countRpm(); c > 0 {
		info.Managers = append(info.Managers, "rpm")
		info.Count += c
	}
	if c := countPacman(); c > 0 {
		info.Managers = append(info.Managers, "pacman")
		info.Count += c
	}
	if c := countDirEntries("/var/lib/flatpak/app"); c > 0 {
		info.Managers = append(info.Managers, "flatpak")
		info.Count += c
	}
	if c := countDirEntries("/var/lib/snapd/snaps"); c > 0 {
		info.Managers = append(info.Managers, "snap")
		info.Count += c
	}
	if c := countNix(); c > 0 {
		info.Managers = append(info.Managers, "nix")
		info.Count += c
	}

	return info
}

func countDpkg() int {
	out, err := exec.Command("dpkg-query", "-f", "${Package}\n", "-W").Output()
	if err != nil {
		return 0
	}
	return len(strings.Split(strings.TrimSpace(string(out)), "\n"))
}

func countRpm() int {
	out, err := exec.Command("rpm", "-qa", "--queryformat", "%{name}\n").Output()
	if err != nil {
		return 0
	}
	return len(strings.Split(strings.TrimSpace(string(out)), "\n"))
}

func countPacman() int {
	return countDirEntries("/var/lib/pacman/local")
}

func countNix() int {
	out, err := exec.Command("nix-store", "-q", "--references", "/run/current-system/sw").Output()
	if err != nil {
		return countDirEntries("/nix/store")
	}
	return len(strings.Split(strings.TrimSpace(string(out)), "\n"))
}

func countDirEntries(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
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

	if info.Name == "" {
		if out, err := exec.Command("wmctrl", "-m").Output(); err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Name:") {
					info.Name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
					break
				}
			}
		}
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessionType := os.Getenv("XDG_SESSION_TYPE")

	if sessionType == "wayland" {
		if out, err := exec.CommandContext(ctx, "wlr-randr").Output(); err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "Enabled") || strings.Contains(line, "current") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						info.Resolutions = append(info.Resolutions, parts[0])
					}
				}
			}
		}
		if len(info.Resolutions) > 0 {
			return info
		}
	}

	displayVar := os.Getenv("DISPLAY")
	if displayVar == "" {
		return info
	}

	cmd := exec.CommandContext(ctx, "xrandr")
	cmd.Env = []string{
		"DISPLAY=" + displayVar,
		"PATH=" + os.Getenv("PATH"),
	}
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

	var totalCurrent, totalFull float64
	var statuses []string

	for _, b := range batteries {
		totalCurrent += b.Current
		totalFull += b.Full
		switch b.State.Raw {
		case battery.Charging:
			statuses = append(statuses, "Charging")
		case battery.Discharging:
			statuses = append(statuses, "Discharging")
		case battery.Full:
			statuses = append(statuses, "Full")
		case battery.Idle:
			statuses = append(statuses, "AC Connected")
		}
	}

	info := &BatteryInfo{}
	if totalFull > 0 {
		info.Percentage = int(totalCurrent / totalFull * 100)
	}
	if len(batteries) == 1 || allSame(statuses) {
		if len(statuses) > 0 {
			info.Status = statuses[0]
		}
	} else if len(statuses) > 0 {
		info.Status = statuses[0]
	}

	return info
}

func allSame(ss []string) bool {
	if len(ss) < 2 {
		return true
	}
	for _, s := range ss[1:] {
		if s != ss[0] {
			return false
		}
	}
	return true
}

func GetLocale() *LocaleInfo {
	info := &LocaleInfo{}

	if lang := os.Getenv("LANG"); lang != "" {
		info.Locale = lang
	}

	return info
}
