package display

import (
	"strings"
	"sync/atomic"
)

var colorDisabled atomic.Bool

func SetColorDisabled(v bool) {
	colorDisabled.Store(v)
}

func ColorDisabled() bool {
	return colorDisabled.Load()
}

var Colors = map[string]string{
	"default":        "\033[0m",
	"black":          "\033[30m",
	"red":            "\033[31m",
	"green":          "\033[32m",
	"yellow":         "\033[33m",
	"blue":           "\033[34m",
	"magenta":        "\033[35m",
	"cyan":           "\033[36m",
	"white":          "\033[37m",
	"bright_black":   "\033[90m",
	"bright_red":     "\033[91m",
	"bright_green":   "\033[92m",
	"bright_yellow":  "\033[93m",
	"bright_blue":    "\033[94m",
	"bright_magenta": "\033[95m",
	"bright_cyan":    "\033[96m",
	"bright_white":   "\033[97m",
}

var resetCode = "\033[0m"

func GetColor(name string) string {
	if code, ok := Colors[name]; ok {
		return code
	}
	return Colors["default"]
}

func Paint(text, colorName string) string {
	if colorDisabled.Load() {
		return text
	}
	var b strings.Builder
	b.WriteString(GetColor(colorName))
	b.WriteString(text)
	b.WriteString(resetCode)
	return b.String()
}

func PaintTitle(text, colorName string) string {
	if colorDisabled.Load() {
		return text
	}
	var b strings.Builder
	b.WriteString("\033[1m")
	b.WriteString(GetColor(colorName))
	b.WriteString(text)
	b.WriteString(resetCode)
	return b.String()
}

func ColorExists(name string) bool {
	_, ok := Colors[name]
	return ok
}
