//go:build windows

package sysinfo

import (
	"os"
	"os/exec"
	"strconv"
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
	cmdArgs := []string{"-NoProfile", "-Command",
		`Get-CimInstance Win32_VideoController | ForEach-Object { $_.Name }`}
	out, err := exec.Command("powershell", cmdArgs...).Output()
	if err != nil {
		cmdArgs = []string{"-NoProfile", "-Command",
			`Get-WmiObject Win32_VideoController | ForEach-Object { $_.Name }`}
		out, err = exec.Command("powershell", cmdArgs...).Output()
	}

	var gpus []*GPUInfo
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				gpus = append(gpus, &GPUInfo{Name: line})
			}
		}
	}

	if len(gpus) == 0 {
		gpus = append(gpus, &GPUInfo{Name: "Unknown"})
	}
	return gpus
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

func GetShell() *ShellInfo {
	info := &ShellInfo{Name: "cmd"}

	if pwsh, ok := os.LookupEnv("POWERSHELL_DISTRIBUTION_CHANNEL"); ok && pwsh != "" {
		info.Name = "pwsh"
	} else if psVer, ok := os.LookupEnv("PSModulePath"); ok && psVer != "" {
		info.Name = "powershell"
	}

	if shell := os.Getenv("COMSPEC"); shell != "" {
		info.Path = shell
	} else {
		info.Path = "cmd.exe"
	}

	if ver, ok := os.LookupEnv("POWERSHELL_VERSION"); ok && ver != "" {
		info.Version = ver
	}

	return info
}

func GetPackages() *PackageInfo {
	info := &PackageInfo{}

	args := []string{"-NoProfile", "-Command", `winget list --count 2>$null | Measure-Object -Line | Select-Object -ExpandProperty Lines`}
	out, err := exec.Command("powershell", args...).Output()
	if err == nil {
		n := strings.TrimSpace(string(out))
		if n != "" {
			if count, errConv := strconv.Atoi(n); errConv == nil && count > 0 {
				info.Managers = append(info.Managers, "winget")
				info.Count += count
			}
		}
	}
	return info
}

func GetDE() *DEInfo {
	return &DEInfo{Name: "Fluent"}
}

func GetWM() *WMInfo {
	return &WMInfo{Name: "DWM"}
}

func GetTerminal() *TerminalInfo {
	term := os.Getenv("TERM_PROGRAM")
	if term == "" {
		term = "Windows Terminal"
	}
	return &TerminalInfo{Name: term}
}

func GetResolution() *ResolutionInfo {
	info := &ResolutionInfo{}
	cmdArgs := []string{"-NoProfile", "-Command",
		`Add-Type -AssemblyName System.Windows.Forms;[System.Windows.Forms.Screen]::AllScreens | ForEach-Object { "$($_.Bounds.Width)x$($_.Bounds.Height)" }`}
	out, err := exec.Command("powershell", cmdArgs...).Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				info.Resolutions = append(info.Resolutions, line)
			}
		}
	}
	return info
}

func GetHost() *HostInfo {
	info := &HostInfo{}
	cmdArgs := []string{"-NoProfile", "-Command",
		`Get-CimInstance Win32_ComputerSystem | Select-Object -ExpandProperty Model`}
	out, err := exec.Command("powershell", cmdArgs...).Output()
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
	info := &BatteryInfo{}
	cmdArgs := []string{"-NoProfile", "-Command",
		`Get-CimInstance Win32_Battery | Select-Object -ExpandProperty EstimatedChargeRemaining`}
	out, err := exec.Command("powershell", cmdArgs...).Output()
	if err == nil {
		pctStr := strings.TrimSpace(string(out))
		if pct, errConv := strconv.Atoi(pctStr); errConv == nil && pct >= 0 {
			info.Percentage = pct
		}
	}

	statusArgs := []string{"-NoProfile", "-Command",
		`Get-CimInstance Win32_Battery | Select-Object -ExpandProperty BatteryStatus`}
	statusOut, err := exec.Command("powershell", statusArgs...).Output()
	if err == nil {
		switch strings.TrimSpace(string(statusOut)) {
		case "1":
			info.Status = "Discharging"
		case "2":
			info.Status = "AC Connected"
		case "6", "7", "8", "9", "10", "11":
			info.Status = "Charging"
		}
	}
	return info
}

func GetLocale() *LocaleInfo {
	info := &LocaleInfo{}
	cmdArgs := []string{"-NoProfile", "-Command",
		`Get-Culture | Select-Object -ExpandProperty Name`}
	out, err := exec.Command("powershell", cmdArgs...).Output()
	if err == nil {
		info.Locale = strings.TrimSpace(string(out))
	}
	return info
}
