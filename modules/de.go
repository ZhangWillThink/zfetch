package modules

import (
	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&DEModule{})
}

type DEModule struct{}

func (m *DEModule) Name() string { return "de" }

func (m *DEModule) Run() []ModuleInfo {
	info := sysinfo.GetDE()
	if info.Name == "" {
		return nil
	}
	return []ModuleInfo{{Key: "DE", Value: info.Name}}
}
