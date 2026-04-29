package modules

import (
	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&LocalIPModule{})
}

type LocalIPModule struct{}

func (m *LocalIPModule) Name() string { return "localip" }

func (m *LocalIPModule) Run() []ModuleInfo {
	info := sysinfo.GetLocalIP()
	if len(info.Interfaces) == 0 {
		return nil
	}

	var result []ModuleInfo
	for i, entry := range info.Interfaces {
		key := "Local IP"
		if i > 0 {
			key = ""
		}
		value := entry.IP
		if entry.Name != "" {
			value = entry.IP + " (" + entry.Name + ")"
		}
		result = append(result, ModuleInfo{Key: key, Value: value})
	}

	return result
}
