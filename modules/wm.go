package modules

import (
	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&WMModule{})
}

type WMModule struct{}

func (m *WMModule) Name() string { return "wm" }

func (m *WMModule) Run() []ModuleInfo {
	info := sysinfo.GetWM()
	if info.Name == "" {
		return nil
	}
	return []ModuleInfo{{Key: "WM", Value: info.Name}}
}
