package modules

import (
	"strings"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&ResolutionModule{})
}

type ResolutionModule struct{}

func (m *ResolutionModule) Name() string { return "resolution" }

func (m *ResolutionModule) Run() []ModuleInfo {
	info := sysinfo.GetResolution()
	if len(info.Resolutions) == 0 {
		return nil
	}
	return []ModuleInfo{{Key: "Resolution", Value: strings.Join(info.Resolutions, ", ")}}
}
