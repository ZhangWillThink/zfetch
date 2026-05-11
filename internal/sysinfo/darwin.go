//go:build darwin

package sysinfo

import (
	"context"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

func GetOS() *OSInfo {
	info := &OSInfo{Name: "macOS"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "sw_vers", "-productName").Output()
	if err == nil {
		info.Name = strings.TrimSpace(string(out))
	}
	out, err = exec.CommandContext(ctx, "sw_vers", "-productVersion").Output()
	if err == nil {
		info.Version = strings.TrimSpace(string(out))
	}
	info.ID = "macos"
	return info
}

func GetKernel() *KernelInfo {
	h, err := host.Info()
	if err != nil {
		return &KernelInfo{Name: "Darwin"}
	}
	return &KernelInfo{
		Name:    "Darwin",
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "system_profiler", "SPDisplaysDataType").Output()
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

	var deduped []*GPUInfo
	seenName := map[string]bool{}
	for _, g := range gpus {
		key := strings.ToLower(strings.TrimSpace(g.Name))
		if key == "" || seenName[key] {
			continue
		}
		seenName[key] = true
		deduped = append(deduped, g)
	}
	gpus = deduped

	if len(gpus) == 0 {
		gpus = append(gpus, &GPUInfo{Name: "Unknown"})
	}
	return gpus
}

func GetDisk() []*DiskInfo {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	byDev := make(map[uint64]*DiskInfo)
	var noStat []*DiskInfo
	seenPath := map[string]bool{}

	for _, p := range partitions {
		if seenPath[p.Mountpoint] {
			continue
		}
		seenPath[p.Mountpoint] = true

		if IsExcludedFilesystem(p.Fstype) {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		di := &DiskInfo{
			Path:       p.Mountpoint,
			Total:      usage.Total,
			Used:       usage.Used,
			Available:  usage.Free,
			Filesystem: p.Fstype,
		}

		if dev, ok := statMountDev(p.Mountpoint); ok {
			if prev, dup := byDev[dev]; !dup || preferMount(prev.Path, di.Path) {
				byDev[dev] = di
			}
			continue
		}
		noStat = append(noStat, di)
	}

	disks := make([]*DiskInfo, 0, len(byDev)+len(noStat))
	for _, di := range byDev {
		disks = append(disks, di)
	}
	disks = append(disks, noStat...)

	sort.Slice(disks, func(i, j int) bool { return disks[i].Path < disks[j].Path })

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

func GetShell() *ShellInfo {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/zsh"
	}
	parts := strings.Split(shellPath, "/")
	return &ShellInfo{
		Name:    parts[len(parts)-1],
		Path:    shellPath,
		Version: probeShellDashC(shellPath),
	}
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
	name, version := TerminalFromEnv()
	return &TerminalInfo{
		Name:    terminalNameOrUnknown(name),
		Version: version,
	}
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
		info.Percentage = int(float64(b.Current) / float64(b.Full) * 100)
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

func GetLocale() *LocaleInfo {
	return &LocaleInfo{Locale: LocaleFromUnixEnv()}
}
