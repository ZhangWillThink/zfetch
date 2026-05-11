//go:build windows

package sysinfo

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

func GetOS() *OSInfo {
	info := &OSInfo{Name: "Windows"}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "cmd", "/c", "ver").Output()
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

var (
	winScreenHostLocaleOnce sync.Once
	winScreenHostLocale     struct {
		Resolutions []string
		Host        string
		Locale      string
	}
)

func probeWinScreenHostLocale() {
	winScreenHostLocaleOnce.Do(func() {
		ps := findPowerShellExe()
		if ps == "" {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		script := `$ErrorActionPreference='SilentlyContinue';` +
			`Add-Type -AssemblyName System.Windows.Forms;` +
			`$res=@([System.Windows.Forms.Screen]::AllScreens | ForEach-Object { "$($_.Bounds.Width)x$($_.Bounds.Height)" });` +
			`$h=''; try { $h=(Get-CimInstance Win32_ComputerSystem).Model.Trim() } catch {} ;` +
			`$loc=''; try { $loc=(Get-Culture).Name.Trim() } catch {} ;` +
			`@{ res=$res; host=$h; loc=$loc } | ConvertTo-Json -Compress -Depth 6`
		out, err := exec.CommandContext(ctx, ps, "-NoProfile", "-Command", script).Output()
		if err != nil {
			return
		}
		var dec struct {
			Res  []string `json:"res"`
			Host string   `json:"host"`
			Loc  string   `json:"loc"`
		}
		decJSON := strings.TrimSpace(string(bytes.TrimPrefix(bytes.TrimSpace(out), []byte{0xef, 0xbb, 0xbf})))
		if json.Unmarshal([]byte(decJSON), &dec) != nil {
			return
		}
		winScreenHostLocale.Resolutions = dec.Res
		winScreenHostLocale.Host = strings.TrimSpace(dec.Host)
		winScreenHostLocale.Locale = strings.TrimSpace(dec.Loc)
	})
}

func GetGPU() []*GPUInfo {
	ps := findPowerShellExe()
	if ps == "" {
		return []*GPUInfo{{Name: "Unknown"}}
	}
	cmdArgs := []string{"-NoProfile", "-Command",
		`Get-CimInstance Win32_VideoController | ForEach-Object { $_.Name }`}
	out, err := exec.Command(ps, cmdArgs...).Output()
	if err != nil {
		cmdArgs = []string{"-NoProfile", "-Command",
			`Get-WmiObject Win32_VideoController | ForEach-Object { $_.Name }`}
		out, err = exec.Command(ps, cmdArgs...).Output()
	}

	var raw []string
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				raw = append(raw, line)
			}
		}
	}

	var gpus []*GPUInfo
	have := map[string]bool{}
	for _, name := range raw {
		key := strings.ToLower(name)
		if have[key] {
			continue
		}
		have[key] = true
		gpus = append(gpus, &GPUInfo{Name: name})
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

	byDevice := make(map[string]*DiskInfo)
	var noDev []*DiskInfo
	seenMount := map[string]bool{}

	for _, p := range partitions {
		if seenMount[p.Mountpoint] {
			continue
		}
		seenMount[p.Mountpoint] = true

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

		devKey := normalizeWindowsVolDevice(p.Device)
		if devKey != "" {
			if prev, ok := byDevice[devKey]; !ok || windowsPreferMount(prev.Path, di.Path) {
				byDevice[devKey] = di
			}
			continue
		}
		noDev = append(noDev, di)
	}

	disks := make([]*DiskInfo, 0, len(byDevice)+len(noDev))
	for _, di := range byDevice {
		disks = append(disks, di)
	}
	disks = append(disks, noDev...)

	sort.Slice(disks, func(i, j int) bool {
		return strings.ToLower(disks[i].Path) < strings.ToLower(disks[j].Path)
	})

	if len(disks) == 0 {
		disks = append(disks, &DiskInfo{Path: "C:\\"})
	}
	return disks
}

func normalizeWindowsVolDevice(device string) string {
	d := strings.TrimSpace(device)
	d = strings.TrimPrefix(d, `\\.\`)
	if d == "" {
		return ""
	}
	return strings.ToUpper(d)
}

func windowsPreferMount(cur, cand string) bool {
	cu := strings.ToUpper(cur)
	cc := strings.ToUpper(cand)
	if strings.HasPrefix(cc, `C:\`) && !strings.HasPrefix(cu, `C:\`) {
		return true
	}
	if strings.HasPrefix(cu, `C:\`) && !strings.HasPrefix(cc, `C:\`) {
		return false
	}
	return len(cc) < len(cu)
}

func findPowerShellExe() string {
	for _, name := range []string{"pwsh.exe", "pwsh", "powershell.exe", "powershell"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

func GetShell() *ShellInfo {
	info := &ShellInfo{}
	shellPath := strings.TrimSpace(os.Getenv("SHELL"))
	if shellPath != "" {
		base := strings.ToLower(strings.TrimSuffix(filepath.Base(shellPath), ".exe"))
		switch base {
		case "bash", "zsh", "fish", "sh":
			info.Path = shellPath
			info.Name = filepath.Base(shellPath)
			info.Version = probeShellDashC(shellPath)
			return info
		}
	}

	if psPath := findPowerShellExe(); psPath != "" {
		info.Path = psPath
		info.Name = strings.TrimSuffix(strings.ToLower(filepath.Base(psPath)), ".exe")
		info.Version = probePowerShellVersion(psPath)
		return info
	}

	comspec := os.Getenv("COMSPEC")
	if comspec == "" {
		comspec = `C:\Windows\System32\cmd.exe`
	}
	info.Path = comspec
	info.Name = filepath.Base(comspec)
	return info
}

func GetPackages() *PackageInfo {
	info := &PackageInfo{}
	ps := findPowerShellExe()
	if ps == "" {
		return info
	}
	args := []string{"-NoProfile", "-Command", `winget list --count 2>$null | Measure-Object -Line | Select-Object -ExpandProperty Lines`}
	out, err := exec.Command(ps, args...).Output()
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
	name, version := TerminalFromEnv()
	if strings.TrimSpace(name) == "" && os.Getenv("WT_SESSION") != "" {
		name = "Windows Terminal"
	}
	return &TerminalInfo{Name: terminalNameOrUnknown(name), Version: version}
}

func GetResolution() *ResolutionInfo {
	probeWinScreenHostLocale()
	info := &ResolutionInfo{}
	for _, line := range winScreenHostLocale.Resolutions {
		line = strings.TrimSpace(line)
		if line != "" {
			info.Resolutions = append(info.Resolutions, line)
		}
	}
	return info
}

func GetHost() *HostInfo {
	probeWinScreenHostLocale()
	return &HostInfo{Product: winScreenHostLocale.Host}
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
	ps := findPowerShellExe()
	if ps == "" {
		return info
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	script := `$ErrorActionPreference='SilentlyContinue';` +
		`$b=(Get-CimInstance Win32_Battery -ErrorAction SilentlyContinue) | Select-Object -First 1;` +
		`if(-not $b){ @{ pct=-1; st='' } | ConvertTo-Json -Compress; return };` +
		`$pct=[int]$b.EstimatedChargeRemaining; $bs=[int]$b.BatteryStatus; $st='';` +
		`switch($bs){1{$st='Discharging'}2{$st='AC Connected'} {(6..11)-contains$_}{$st='Charging'} };` +
		`@{ pct=$pct; st=$st } | ConvertTo-Json -Compress`
	out, err := exec.CommandContext(ctx, ps, "-NoProfile", "-Command", script).Output()
	if err != nil {
		return info
	}
	var dec struct {
		Pct int    `json:"pct"`
		St  string `json:"st"`
	}
	raw := bytes.TrimSpace(out)
	raw = bytes.TrimPrefix(raw, []byte{0xef, 0xbb, 0xbf})
	if json.Unmarshal(raw, &dec) != nil {
		return info
	}
	if dec.Pct >= 0 && dec.Pct <= 100 {
		info.Percentage = dec.Pct
	}
	info.Status = dec.St
	return info
}

func GetLocale() *LocaleInfo {
	probeWinScreenHostLocale()
	info := &LocaleInfo{Locale: winScreenHostLocale.Locale}
	info.Locale = strings.TrimSpace(info.Locale)
	if info.Locale == "" {
		info.Locale = LocaleFromUnixEnv()
	}
	return info
}
