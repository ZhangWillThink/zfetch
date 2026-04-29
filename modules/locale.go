package modules

import (
	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&LocaleModule{})
}

type LocaleModule struct{}

func (m *LocaleModule) Name() string { return "locale" }

func (m *LocaleModule) Run() []ModuleInfo {
	info := sysinfo.GetLocale()
	if info.Locale == "" {
		return nil
	}

	return []ModuleInfo{{Key: "Locale", Value: info.Locale}}
}
