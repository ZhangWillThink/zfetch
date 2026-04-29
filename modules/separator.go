package modules

import (
	"strings"
)

func init() {
	Register(&SeparatorModule{})
}

type SeparatorModule struct{}

func (m *SeparatorModule) Name() string { return "separator" }

func (m *SeparatorModule) Run() []ModuleInfo {
	return []ModuleInfo{{Key: strings.Repeat("-", 30), Value: ""}}
}
