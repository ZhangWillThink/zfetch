//go:build linux

package sysinfo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	drmCardNumRe = regexp.MustCompile(`^card\d+$`)
	pciDomRe     = regexp.MustCompile(`(?i)^[\da-f]{4}:[\da-f]{2}:[\da-f]{2}\.\d+$`)
	pciBusRe     = regexp.MustCompile(`(?i)^[\da-f]{2}:[\da-f]{2}\.\d+$`)
)

func GetGPU() []*GPUInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lspciRows := listPCIeDisplayControllers(ctx)
	drmRows := enumerateDRMDevices()

	drmByPCI := make(map[string]gpuDRMEntry)
	for _, d := range drmRows {
		if d.pci != "" {
			drmByPCI[d.pci] = d
		}
	}

	pciSeen := map[string]bool{}

	var merged []*GPUInfo

	if len(lspciRows) > 0 {
		for _, row := range lspciRows {
			g := &GPUInfo{Name: row.name}
			if ex, ok := drmByPCI[row.slot]; ok {
				fillGPUFromDRM(g, ex)
			}
			if row.slot != "" {
				pciSeen[row.slot] = true
			}
			merged = append(merged, g)
		}
	}

	for _, d := range drmRows {
		if d.pci != "" && pciSeen[d.pci] {
			continue
		}
		name := strings.TrimSpace(d.sysfsName)
		if name == "" {
			continue
		}
		g := &GPUInfo{Name: name}
		fillGPUFromDRM(g, d)
		if d.pci != "" {
			pciSeen[d.pci] = true
		}
		merged = append(merged, g)
	}

	if len(merged) == 0 {
		for _, row := range lspciRows {
			merged = append(merged, &GPUInfo{Name: row.name})
		}
	}

	if len(merged) == 0 {
		return nil
	}

	return pruneVirtualGPUs(merged)
}

type gpuLspci struct {
	slot string
	name string
}

func listPCIeDisplayControllers(ctx context.Context) []gpuLspci {
	cmd := exec.CommandContext(ctx, "lspci", "-mm")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var rows []gpuLspci
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ul := strings.ToUpper(line)
		if !strings.Contains(ul, "VGA") && !strings.Contains(ul, "3D") &&
			!strings.Contains(ul, "DISPLAY") {
			continue
		}

		fields := strings.Split(line, `"`)
		if len(fields) < 2 || strings.TrimSpace(fields[0]) == "" {
			continue
		}
		rawSlot := strings.TrimSpace(fields[0])

		vendor := ""
		device := ""
		if len(fields) > 5 {
			vendor = strings.TrimSpace(fields[3])
			device = strings.TrimSpace(fields[5])
		} else if len(fields) >= 4 {
			vendor = strings.TrimSpace(fields[3])
		}
		label := gpuVendorProductLine(vendor, device)
		if label == "" {
			continue
		}
		rows = append(rows, gpuLspci{
			slot: normalizePCISlot(rawSlot),
			name: label,
		})
	}
	return rows
}

func gpuVendorProductLine(vendor, product string) string {
	vendor = strings.TrimSpace(vendor)
	product = strings.TrimSpace(product)
	if vendor == "" {
		return product
	}
	if product == "" {
		return vendor
	}
	pl, vl := strings.ToLower(product), strings.ToLower(vendor)
	if strings.Contains(pl, vl) {
		return product
	}
	return vendor + " " + product
}

type gpuDRMEntry struct {
	pci        string
	sysfsName  string
	vendorID   string
	memMiB     int
	gpuType    string
	driverHint string
}

func enumerateDRMDevices() []gpuDRMEntry {
	drmRoot := "/sys/class/drm"
	ents, err := os.ReadDir(drmRoot)
	if err != nil {
		return nil
	}

	pciDup := map[string]bool{}
	var out []gpuDRMEntry

	for _, ent := range ents {
		base := ent.Name()
		if !drmCardNumRe.MatchString(base) {
			continue
		}
		cardRoot := filepath.Join(drmRoot, base)
		devPath := filepath.Join(cardRoot, "device")

		vdat, err := os.ReadFile(filepath.Join(devPath, "vendor"))
		vendorID := strings.TrimSpace(string(vdat))
		if err != nil || vendorID == "" {
			continue
		}

		drmName := parseGPUDeviceNames(cardRoot)
		if drmName == "" {
			continue
		}

		pci := pciSlotFromDeviceDir(devPath)
		if pci != "" && pciDup[pci] {
			continue
		}
		if pci != "" {
			pciDup[pci] = true
		}

		memMiB, role := readVRAMAndGPURole(devPath)
		driver := readDRMDeviceDriver(devPath)

		out = append(out, gpuDRMEntry{
			pci:        pci,
			sysfsName:  strings.TrimSpace(drmName),
			vendorID:   vendorID,
			memMiB:     memMiB,
			gpuType:    role,
			driverHint: driver,
		})
	}
	return out
}

func readDRMDeviceDriver(devPath string) string {
	link, err := filepath.EvalSymlinks(filepath.Join(devPath, "driver"))
	if err != nil {
		return ""
	}
	return link
}

func fillGPUFromDRM(g *GPUInfo, e gpuDRMEntry) {
	g.VendorPCIID = e.vendorID
	if e.memMiB > 0 {
		g.MemoryMiB = e.memMiB
	}
	if e.gpuType != "" {
		g.Type = e.gpuType
	}
	if e.driverHint != "" {
		g.Driver = filepath.Base(e.driverHint)
	}
}

func readVRAMAndGPURole(devPath string) (memMiB int, role string) {
	if vramData, err := os.ReadFile(filepath.Join(devPath, "mem_info_vram_total")); err == nil {
		b, _ := strconv.ParseUint(strings.TrimSpace(string(vramData)), 10, 64)
		memMiB = int(b / (1024 * 1024))
	}

	if bootData, err := os.ReadFile(filepath.Join(devPath, "boot_vga")); err == nil {
		if strings.TrimSpace(string(bootData)) == "1" {
			role = "Discrete"
		}
	}
	if role == "" {
		driverPath := filepath.Join(devPath, "driver")
		if resolved, err := filepath.EvalSymlinks(driverPath); err == nil {
			rs := resolved
			if strings.Contains(rs, "i915") || strings.Contains(rs, "amdgpu/apu") {
				role = "Integrated"
			} else {
				role = "Discrete"
			}
		}
	}
	return memMiB, role
}

func parseGPUDeviceNames(cardRoot string) string {
	ents, err := os.ReadDir(filepath.Join(cardRoot, "device"))
	if err != nil {
		return ""
	}
	for _, e := range ents {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "drm") {
			continue
		}
		drmSub := filepath.Join(cardRoot, "device", e.Name())
		vendorData, _ := os.ReadFile(filepath.Join(drmSub, "vendor_name"))
		deviceData, _ := os.ReadFile(filepath.Join(drmSub, "device_name"))
		vendor := strings.TrimSpace(string(vendorData))
		dev := strings.TrimSpace(string(deviceData))
		if vendor != "" || dev != "" {
			return strings.TrimSpace(vendor + " " + dev)
		}
	}
	return ""
}

func pciSlotFromDeviceDir(devicePath string) string {
	link, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return ""
	}
	return extractPCISegment(link)
}

func extractPCISegment(resolvedPath string) string {
	for dir := filepath.Clean(resolvedPath); strings.HasPrefix(dir, string(filepath.Separator)); {
		base := filepath.Base(dir)
		if pciDomRe.MatchString(base) {
			return normalizePCISlot(base)
		}
		if pciBusRe.MatchString(base) {
			return normalizePCISlot(base)
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	return ""
}

func normalizePCISlot(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	switch {
	case pciDomRe.MatchString(s):
		return s
	case pciBusRe.MatchString(s):
		return "0000:" + s
	default:
		return s
	}
}

func pruneVirtualGPUs(gs []*GPUInfo) []*GPUInfo {
	if len(gs) == 0 {
		return gs
	}

	hasNonVirt := false
	for _, g := range gs {
		v := strings.ToLower(strings.TrimSpace(g.VendorPCIID))
		if v != "" && !isVirtualGPUVendorPCI(v) {
			hasNonVirt = true
			break
		}
		if v == "" && looksLikeRealDiscreteName(g.Name) {
			hasNonVirt = true
			break
		}
	}

	if !hasNonVirt {
		return dedupeByName(gs)
	}

	out := gs[:0]
	for _, g := range gs {
		v := strings.ToLower(strings.TrimSpace(g.VendorPCIID))
		if v != "" && isVirtualGPUVendorPCI(v) {
			continue
		}

		out = append(out, g)
	}
	if len(out) == 0 {
		return dedupeByName(gs)
	}
	return out
}

func isVirtualGPUVendorPCI(v string) bool {
	switch strings.TrimSpace(v) {
	case "0x1414", "0x1af4", "0x1234":
		return true
	default:
		return false
	}
}

func looksLikeRealDiscreteName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return false
	}
	valid := strings.Contains(n, "nvidia") ||
		strings.Contains(n, "geforce") ||
		strings.Contains(n, "radeon") ||
		strings.Contains(n, "rx ") ||
		strings.Contains(n, "rtx") ||
		strings.Contains(n, "gtx") ||
		strings.Contains(n, "intel") ||
		strings.Contains(n, "iris") ||
		strings.Contains(n, "arc") ||
		strings.Contains(n, "uhd ") ||
		strings.Contains(n, "integrated graphics")
	return valid &&
		!strings.Contains(n, "virtio") &&
		!strings.HasPrefix(n, "microsoft corporation")
}

func dedupeByName(gs []*GPUInfo) []*GPUInfo {
	seen := map[string]bool{}
	out := []*GPUInfo{}
	for _, g := range gs {
		key := strings.ToLower(strings.TrimSpace(g.Name))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, g)
	}
	if len(out) == 0 {
		return gs
	}
	return out
}
