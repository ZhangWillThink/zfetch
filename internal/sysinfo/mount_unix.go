//go:build unix

package sysinfo

import (
	"strings"
	"syscall"
)

// statMountDev returns Linux/Darwin filesystem device id from stat(path).
// Duplicate bind mounts typically share this value.
func statMountDev(path string) (dev uint64, ok bool) {
	var st syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		return 0, false
	}
	return uint64(st.Dev), true
}

// preferMount replaces current with cand when cand should be shown instead
// when both refer to the same underlying device (duplicate bind mounts).
func preferMount(curPath, candPath string) bool {
	if candPath == "/" {
		return true
	}
	if curPath == "/" {
		return false
	}
	curWSL := strings.HasPrefix(curPath, "/mnt/wslg/")
	candWSL := strings.HasPrefix(candPath, "/mnt/wslg/")
	if curWSL && !candWSL {
		return false
	}
	if !curWSL && candWSL {
		return true
	}
	curDist := strings.HasPrefix(curPath, "/mnt/wslg/distro")
	candDist := strings.HasPrefix(candPath, "/mnt/wslg/distro")
	if curDist && !candDist {
		return false
	}
	if !curDist && candDist {
		return true
	}
	if lc, lcc := len(curPath), len(candPath); lc != lcc {
		return lcc < lc
	}
	return candPath < curPath
}
