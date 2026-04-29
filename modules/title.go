package modules

import (
	"os"
	"os/user"
)

func init() {
	Register(&TitleModule{})
}

type TitleModule struct{}

func (m *TitleModule) Name() string { return "title" }

func (m *TitleModule) Run() []ModuleInfo {
	userName := ""
	if u, err := user.Current(); err == nil {
		userName = u.Username
	}
	if userName == "" {
		userName = os.Getenv("USER")
	}

	hostname, _ := os.Hostname()

	if userName != "" && hostname != "" {
		return []ModuleInfo{{Key: userName + "@" + hostname, Value: ""}}
	} else if userName != "" {
		return []ModuleInfo{{Key: userName, Value: ""}}
	} else if hostname != "" {
		return []ModuleInfo{{Key: hostname, Value: ""}}
	}
	return []ModuleInfo{{Key: "unknown", Value: ""}}
}
