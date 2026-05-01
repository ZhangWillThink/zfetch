package modules

import (
	"fmt"
	"strings"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&PackagesModule{})
}

type PackagesModule struct{}

func (m *PackagesModule) Name() string { return "packages" }

func (m *PackagesModule) Run() []ModuleInfo {
	info := sysinfo.GetPackages()
	if info.Count == 0 {
		return []ModuleInfo{{Key: "Packages", Value: "Unknown"}}
	}
	var val string
	if len(info.Managers) > 0 {
		val = fmt.Sprintf("%d (%s)", info.Count, strings.Join(info.Managers, ", "))
	} else {
		val = fmt.Sprintf("%d", info.Count)
	}
	return []ModuleInfo{{Key: "Packages", Value: val}}
}
