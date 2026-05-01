//go:build windows

package sysinfo

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

func GetOS() *OSInfo {
	info := &OSInfo{Name: "Windows"}
	out, err := exec.Command("cmd", "/c", "ver").Output()
	if err == nil {
		info.Version = strings.TrimSpace(string(out))
	}
	info.ID = "windows"
	return info
}

func GetKernel() *KernelInfo {
	h, err := host.Info()
	if err != nil {
		return &KernelInfo{Name: "Windows NT"}
	}
	return &KernelInfo{
		Name:    "Windows NT",
		Release: h.KernelVersion,
		Machine: h.KernelArch,
	}
}

func GetCPU() *CPUInfo {
	cpus, err := cpu.Info()
	if err != nil || len(cpus) == 0 {
		info := &CPUInfo{}
		info.Name = os.Getenv("PROCESSOR_IDENTIFIER")
		if info.Name == "" {
			info.Name = "Unknown"
		}
		return info
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
	info := &GPUInfo{}
	out, err := exec.Command("wmic", "path", "win32_videocontroller", "get", "name").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			info.Name = strings.TrimSpace(lines[1])
		}
	}
	if info.Name == "" {
		info.Name = "Unknown"
	}
	return []*GPUInfo{info}
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
		return []*DiskInfo{{Path: "C:\\"}}
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

	if len(disks) == 0 {
		disks = append(disks, &DiskInfo{Path: "C:\\"})
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
	shell := os.Getenv("COMSPEC")
	if shell == "" {
		shell = "cmd.exe"
	}
	return &ShellInfo{Name: "cmd", Path: shell}
}

func GetPackages() *PackageInfo {
	return &PackageInfo{}
}

func GetDE() *DEInfo {
	return &DEInfo{}
}

func GetWM() *WMInfo {
	return &WMInfo{}
}

func GetTerminal() *TerminalInfo {
	term := os.Getenv("TERM_PROGRAM")
	if term == "" {
		term = "Windows Terminal"
	}
	return &TerminalInfo{Name: term}
}

func GetResolution() *ResolutionInfo {
	return &ResolutionInfo{}
}

func GetHost() *HostInfo {
	return &HostInfo{}
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
	return &BatteryInfo{}
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
	return &LocaleInfo{}
}
