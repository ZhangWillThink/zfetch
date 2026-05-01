package sysinfo

import (
	"fmt"
	"net"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type OSInfo struct {
	Name    string
	Version string
	ID      string
}

type KernelInfo struct {
	Name    string
	Release string
	Version string
	Machine string
}

type CPUInfo struct {
	Name   string
	Cores  int
	MaxMHz float64
}

type GPUInfo struct {
	Name      string
	Driver    string
	MemoryMiB int
	Type      string
}

type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
}

type DiskInfo struct {
	Total      uint64
	Used       uint64
	Available  uint64
	Path       string
	Filesystem string
}

type PackageInfo struct {
	Count    int
	Managers []string
}

type ShellInfo struct {
	Name    string
	Version string
	Path    string
}

type UptimeInfo struct {
	Uptime uint64
}

type DEInfo struct {
	Name    string
	Version string
}

type WMInfo struct {
	Name string
}

type TerminalInfo struct {
	Name    string
	Version string
}

type ResolutionInfo struct {
	Resolutions []string
}

type HostInfo struct {
	Product string
	Version string
}

type SwapInfo struct {
	Total uint64
	Used  uint64
}

type BatteryInfo struct {
	Percentage int
	Status     string
}

type LocalIPEntry struct {
	Name string
	IP   string
}

type LocalIPInfo struct {
	Interfaces []LocalIPEntry
}

type LocaleInfo struct {
	Locale string
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

func GetUptime() *UptimeInfo {
	u, err := host.Uptime()
	if err != nil {
		return &UptimeInfo{}
	}
	return &UptimeInfo{Uptime: u}
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
