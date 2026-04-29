//go:build windows

package sysinfo

import (
	"os"
	"os/exec"
	"strings"
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
	info := &KernelInfo{Name: "Windows NT"}
	out, err := exec.Command("cmd", "/c", "ver").Output()
	if err == nil {
		info.Release = strings.TrimSpace(string(out))
	}
	return info
}

func GetCPU() *CPUInfo {
	info := &CPUInfo{}
	out, err := exec.Command("wmic", "cpu", "get", "name").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			info.Name = strings.TrimSpace(lines[1])
		}
	}
	if info.Name == "" {
		info.Name = os.Getenv("PROCESSOR_IDENTIFIER")
	}
	return info
}

func GetGPU() *GPUInfo {
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
	return info
}

func GetMemory() *MemoryInfo {
	return &MemoryInfo{}
}

func GetDisk() *DiskInfo {
	return &DiskInfo{Path: "C:\\"}
}

func GetUptime() *UptimeInfo {
	return &UptimeInfo{}
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
