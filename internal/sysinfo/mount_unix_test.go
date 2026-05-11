//go:build unix

package sysinfo

import (
	"testing"
)

func TestPreferMount(t *testing.T) {
	if !preferMount("/mnt/wslg/distro", "/") {
		t.Fatal("prefer / over wsl distro bind")
	}
	if preferMount("/", "/mnt/wslg/distro") {
		t.Fatal("do not swap / for distro path")
	}
	if !preferMount("/media/very/long/mount/name", "/media/a") {
		t.Fatal("prefer shorter mount path tie-break")
	}
}
