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
	Name   string
	Driver string
}

type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
}

type DiskInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
	Path      string
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
