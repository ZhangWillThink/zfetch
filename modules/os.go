package modules

import (
	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&OSModule{})
}

type OSModule struct{}

func (m *OSModule) Name() string { return "os" }

func (m *OSModule) Run() []ModuleInfo {
	info := sysinfo.GetOS()
	result := info.Name
	if info.Version != "" {
		result += " " + info.Version
	}
	return []ModuleInfo{{Key: "OS", Value: result}}
}
