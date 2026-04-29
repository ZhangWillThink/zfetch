package display

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/WillZhang/zfetch/config"
	"github.com/WillZhang/zfetch/modules"
)

type Display struct {
	cfg  *config.Config
	pipe bool
}

func New(cfg *config.Config, pipe bool) *Display {
	if cfg.ColorKeys == "default" {
		cfg.ColorKeys = "bright_green"
	}
	if cfg.ColorTitle == "default" {
		cfg.ColorTitle = "bright_white"
	}
	return &Display{cfg: cfg, pipe: pipe}
}

func (d *Display) Render() {
	structure := d.cfg.Structure

	moduleKeys := strings.Split(structure, ":")
	if len(moduleKeys) == 0 {
		return
	}

	var lineInfos []modules.ModuleInfo

	for _, key := range moduleKeys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		if key == "separator" {
			lineInfos = append(lineInfos, modules.ModuleInfo{Key: "separator", Value: ""})
			continue
		}

		m := modules.Get(key)
		if m == nil {
			lineInfos = append(lineInfos, modules.ModuleInfo{Key: key, Value: "unknown module"})
			continue
		}

		results := m.Run()
		lineInfos = append(lineInfos, results...)
	}

	if d.pipe {
		d.renderPipe(lineInfos)
	} else {
		d.renderWithLogo(lineInfos)
	}
}

func (d *Display) renderPipe(infos []modules.ModuleInfo) {
	for _, info := range infos {
		if info.Key == "" && info.Value == "" {
			fmt.Println()
			continue
		}
		if info.Value == "" {
			fmt.Println(info.Key)
		} else {
			fmt.Printf("%s%s%s\n", info.Key, d.cfg.Separator, info.Value)
		}
	}
}

func (d *Display) renderWithLogo(infos []modules.ModuleInfo) {
	logoName := d.cfg.Logo
	if logoName == "" {
		logoName = detectOSLogo()
	}

	logo := GetLogo(logoName)
	if len(logo) == 0 {
		d.renderPipe(infos)
		return
	}
	logoWidth := maxRuneWidth(logo) + 3

	keyColor := d.cfg.ColorKeys
	titleColor := d.cfg.ColorTitle

	maxKeyWidth := 0
	for _, info := range infos {
		if info.Key == "" || info.Key == "separator" {
			continue
		}
		w := utf8.RuneCountInString(info.Key)
		if w > maxKeyWidth {
			maxKeyWidth = w
		}
	}

	type row struct {
		left  string
		right string
	}

	var rows []row

	n := len(infos)
	if len(logo) > n {
		n = len(logo)
	}

	for i := 0; i < n; i++ {
		var left string
		if i < len(logo) {
			pad := logoWidth - runeWidth(logo[i])
			if pad > 0 {
				left = logo[i] + strings.Repeat(" ", pad)
			} else {
				left = logo[i]
			}
		} else if len(logo) > 0 {
			left = strings.Repeat(" ", logoWidth)
		}

		var right string
		if i < len(infos) {
			info := infos[i]
			if info.Key == "separator" {
				right = strings.Repeat("─", 40)
			} else if info.Value == "" {
				pk := padRunes(info.Key, maxKeyWidth)
				right = pk
			} else {
				pk := padRunes(info.Key, maxKeyWidth)
				right = pk + " " + d.cfg.Separator + " " + info.Value
			}
		}

		rows = append(rows, row{left: left, right: right})
	}

	sep := "  "
	for i, r := range rows {
		if r.left != "" {
			fmt.Print(Paint(r.left, keyColor))
			fmt.Print(sep)
		}

		if r.right == "" {
			fmt.Println()
			continue
		}

		if i < len(infos) && infos[i].Key == "separator" {
			fmt.Print(Paint(r.right, keyColor))
		} else if i == 0 && i < len(infos) && infos[i].Value == "" {
			fmt.Print(PaintTitle(r.right, titleColor))
		} else {
			isTitle := i == 0 && i < len(infos)
			coloredKey, rest := splitColored(r.right, maxKeyWidth, keyColor, isTitle)
			fmt.Print(coloredKey)
			if i < len(infos) && infos[i].UsagePercent > 0 {
				fmt.Print(Paint(rest, usageColor(infos[i].UsagePercent)))
			} else {
				fmt.Print(rest)
			}
		}

		fmt.Println()
	}
}

func splitColored(right string, keyWidth int, color string, isTitle bool) (coloredKey string, rest string) {
	key, after, found := strings.Cut(right, " ")
	if !found {
		if isTitle {
			return PaintTitle(right, color), ""
		}
		return Paint(right, color), ""
	}
	if isTitle {
		return PaintTitle(key, color), " " + after
	}
	return Paint(key, color), " " + after
}

func runeWidth(s string) int {
	return utf8.RuneCountInString(s)
}

func maxRuneWidth(lines []string) int {
	m := 0
	for _, l := range lines {
		w := runeWidth(l)
		if w > m {
			m = w
		}
	}
	return m
}

func usageColor(percent float64) string {
	switch {
	case percent >= 85:
		return "bright_red"
	case percent >= 60:
		return "bright_yellow"
	default:
		return "bright_green"
	}
}

func padRunes(s string, width int) string {
	cur := utf8.RuneCountInString(s)
	if cur >= width {
		return s
	}
	return s + strings.Repeat(" ", width-cur)
}
