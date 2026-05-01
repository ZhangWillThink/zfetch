package display

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"

	"github.com/WillZhang/zfetch/config"
	"github.com/WillZhang/zfetch/modules"
)

type Display struct {
	cfg       *config.Config
	pipe      bool
	termWidth int
}

func New(cfg *config.Config, pipe bool) *Display {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	ccfg := *cfg
	if ccfg.ColorKeys == "default" {
		ccfg.ColorKeys = "bright_green"
	}
	if ccfg.ColorTitle == "default" {
		ccfg.ColorTitle = "bright_white"
	}
	return &Display{cfg: &ccfg, pipe: pipe, termWidth: getTerminalWidth()}
}

type indexedResult struct {
	idx  int
	info []modules.ModuleInfo
}

func (d *Display) Render() {
	structure := d.cfg.Structure

	moduleKeys := strings.Split(structure, ":")
	if len(moduleKeys) == 0 {
		return
	}

	var (
		lineInfos = make([]modules.ModuleInfo, 0, len(moduleKeys))
		resultsCh = make(chan indexedResult, len(moduleKeys))
		wg        sync.WaitGroup
	)

	for idx, key := range moduleKeys {
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

		wg.Add(1)
		go func(i int, mod modules.Module) {
			defer wg.Done()
			resultsCh <- indexedResult{idx: i, info: mod.Run()}
		}(idx, m)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var results []indexedResult
	for r := range resultsCh {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].idx < results[j].idx })
	for _, r := range results {
		lineInfos = append(lineInfos, r.info...)
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
	keyColor := d.cfg.ColorKeys
	titleColor := d.cfg.ColorTitle
	sepLen := runewidth.StringWidth(d.cfg.Separator)

	maxKeyWidth := 0
	longestRight := 40
	for _, info := range infos {
		if info.Key == "" || info.Key == "separator" {
			continue
		}
		w := runewidth.StringWidth(info.Key)
		if w > maxKeyWidth {
			maxKeyWidth = w
		}
		var lineW int
		if info.Value == "" {
			lineW = maxKeyWidth
		} else {
			lineW = maxKeyWidth + 1 + sepLen + 1 + runewidth.StringWidth(info.Value)
		}
		if lineW > longestRight {
			longestRight = lineW
		}
	}

	logoName := d.cfg.Logo
	if logoName == "" {
		logoName = detectOSLogo()
	}

	logo := GetLogo(logoName)
	if len(logo) == 0 {
		d.renderInline(infos)
		return
	}

	logoWidth := 0
	for _, l := range logo {
		if w := runewidth.StringWidth(l); w > logoWidth {
			logoWidth = w
		}
	}
	logoWidth += 3
	totalWidth := logoWidth + 2 + longestRight

	if d.termWidth > 0 && totalWidth > d.termWidth {
		d.renderInline(infos)
		return
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
			pad := logoWidth - runewidth.StringWidth(logo[i])
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
				right = strings.Repeat("─", longestRight)
			} else if info.Value == "" {
				pk := runewidth.FillRight(info.Key, maxKeyWidth)
				right = pk
			} else {
				pk := runewidth.FillRight(info.Key, maxKeyWidth)
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

func (d *Display) renderInline(infos []modules.ModuleInfo) {
	maxKeyWidth := 0
	for _, info := range infos {
		if info.Key == "" || info.Key == "separator" {
			continue
		}
		w := runewidth.StringWidth(info.Key)
		if w > maxKeyWidth {
			maxKeyWidth = w
		}
	}

	keyColor := d.cfg.ColorKeys
	titleColor := d.cfg.ColorTitle

	for i, info := range infos {
		if info.Key == "separator" {
			sepLine := strings.Repeat("─", 40)
			if d.termWidth > 0 && 40 > d.termWidth {
				sepLine = strings.Repeat("─", d.termWidth)
			}
			fmt.Println(Paint(sepLine, keyColor))
			continue
		}

		var right string
		if info.Value == "" {
			right = runewidth.FillRight(info.Key, maxKeyWidth)
		} else {
			pk := runewidth.FillRight(info.Key, maxKeyWidth)
			right = pk + " " + d.cfg.Separator + " " + info.Value
		}

		lineWidth := runewidth.StringWidth(right)
		if d.termWidth > 0 && lineWidth > d.termWidth {
			truncateAt := d.termWidth - 3
			if truncateAt < 1 {
				truncateAt = 1
			}
			right = runewidth.Truncate(right, truncateAt, "...")
		}

		isTitle := i == 0 && info.Value == ""
		if isTitle {
			fmt.Println(PaintTitle(right, titleColor))
		} else {
			coloredKey, rest := splitColored(right, maxKeyWidth, keyColor, false)
			fmt.Print(coloredKey)
			if info.UsagePercent > 0 {
				fmt.Print(Paint(rest, usageColor(info.UsagePercent)))
			} else {
				fmt.Print(rest)
			}
			fmt.Println()
		}
	}
}

func splitColored(right string, _ int, color string, isTitle bool) (coloredKey string, rest string) {
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

func usageColor(percent float64) string {
	switch {
	case percent <= 0:
		return "default"
	case percent >= 85:
		return "bright_red"
	case percent >= 60:
		return "bright_yellow"
	default:
		return "bright_green"
	}
}

func getTerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0
	}
	return w
}
