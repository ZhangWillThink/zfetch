package sysinfo

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
