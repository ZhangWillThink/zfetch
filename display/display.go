package display

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"

	"github.com/WillZhang/zfetch/config"
	"github.com/WillZhang/zfetch/modules"
)

const minSidebarRight = 22

type Display struct {
	cfg  *config.Config
	pipe bool
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
	return &Display{cfg: &ccfg, pipe: pipe}
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

func (d *Display) measureKeyColumn(infos []modules.ModuleInfo) (maxKeyWidth int, longestRight int, sepLen int) {
	sepLen = runewidth.StringWidth(d.cfg.Separator)
	longestRight = 40
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
	return maxKeyWidth, longestRight, sepLen
}

func (d *Display) renderWithLogo(infos []modules.ModuleInfo) {
	keyColor := d.cfg.ColorKeys
	titleColor := d.cfg.ColorTitle
	maxKeyWidth, longestRight, sepLen := d.measureKeyColumn(infos)

	logoName := d.cfg.Logo
	if logoName == "" {
		logoName = detectOSLogo()
	}

	logo := GetLogo(logoName)
	if len(logo) == 0 {
		d.renderInline(infos)
		return
	}

	logoPadW := 0
	for _, l := range logo {
		if w := runewidth.StringWidth(l); w > logoPadW {
			logoPadW = w
		}
	}
	logoPadW += 3

	tw := getTerminalWidth()

	if tw <= 0 {
		d.renderLogoSidebarUnbounded(infos, logo, logoPadW, maxKeyWidth, longestRight, sepLen, keyColor, titleColor)
		return
	}

	gaps := []string{"  ", " ", ""}
	for _, gap := range gaps {
		gapLen := runewidth.StringWidth(gap)
		if logoPadW+gapLen > tw {
			break
		}
		rightBudget := tw - logoPadW - gapLen
		if rightBudget < minSidebarRight {
			continue
		}
		prefixW := maxKeyWidth + 1 + sepLen + 1
		if rightBudget <= prefixW+6 {
			continue
		}
		flat, ok := d.buildLogoSidebarFlat(infos, maxKeyWidth, sepLen, rightBudget)
		if !ok || !sidebarFitsWidth(tw, logoPadW, gapLen, flat) {
			continue
		}
		d.printLogoSidebar(infos, logo, logoPadW, gap, flat, maxKeyWidth, keyColor, titleColor)
		return
	}

	if logoPadW+2+longestRight <= tw {
		d.renderLogoSidebarUnbounded(infos, logo, logoPadW, maxKeyWidth, longestRight, sepLen, keyColor, titleColor)
		return
	}

	d.renderInline(infos)
}

// renderLogoSidebarUnbounded mirrors the classic 1:1 logo / info rows (no wrapping).
func (d *Display) renderLogoSidebarUnbounded(infos []modules.ModuleInfo, logo []string, logoPadW, maxKeyWidth, longestRight, sepLen int, keyColor, titleColor string) {
	sep := "  "
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
			pad := logoPadW - runewidth.StringWidth(logo[i])
			if pad > 0 {
				left = logo[i] + strings.Repeat(" ", pad)
			} else {
				left = logo[i]
			}
		} else if len(logo) > 0 {
			left = strings.Repeat(" ", logoPadW)
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
		} else if i == 0 && i < len(infos) && infos[i].Value == "" && infos[i].Key != "" {
			fmt.Print(PaintTitle(r.right, titleColor))
		} else {
			isTitle := i == 0 && i < len(infos) && infos[i].Value == ""
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

type flatSidebarRow struct {
	infoIdx int
	right   string
	cont    bool
	blank   bool
}

func (d *Display) buildLogoSidebarFlat(infos []modules.ModuleInfo, maxKeyWidth, sepLen, avail int) ([]flatSidebarRow, bool) {
	prefixW := maxKeyWidth + 1 + sepLen + 1
	if avail <= prefixW {
		return nil, false
	}

	var rows []flatSidebarRow
	for i, inf := range infos {
		switch {
		case inf.Key == "" && inf.Value == "":
			rows = append(rows, flatSidebarRow{infoIdx: i, blank: true})
		case inf.Key == "separator":
			rows = append(rows, flatSidebarRow{infoIdx: i, right: strings.Repeat("─", avail)})
		case inf.Value == "":
			chunks := wrapParagraph(inf.Key, avail)
			if len(chunks) == 0 {
				chunks = []string{inf.Key}
			}
			for j, ln := range chunks {
				right := ln
				if len(chunks) == 1 && j == 0 {
					right = runewidth.FillRight(ln, maxKeyWidth)
				}
				if runewidth.StringWidth(right) > avail {
					return nil, false
				}
				rows = append(rows, flatSidebarRow{infoIdx: i, right: right, cont: j > 0})
			}
		default:
			pk := runewidth.FillRight(inf.Key, maxKeyWidth)
			head := pk + " " + d.cfg.Separator + " "
			headW := runewidth.StringWidth(head)
			if headW >= avail {
				return nil, false
			}
			firstBudget := avail - headW
			chunks := wrapParagraph(inf.Value, firstBudget)
			if len(chunks) == 0 {
				chunks = []string{inf.Value}
			}
			indent := strings.Repeat(" ", prefixW)
			contAvail := avail - prefixW
			if contAvail <= 0 {
				return nil, false
			}

			var lines []string
			lines = append(lines, head+chunks[0])
			if runewidth.StringWidth(lines[0]) > avail {
				return nil, false
			}
			for _, c := range chunks[1:] {
				parts := wrapParagraph(c, contAvail)
				if len(parts) == 0 {
					parts = []string{c}
				}
				for _, p := range parts {
					ln := indent + p
					if runewidth.StringWidth(ln) > avail {
						return nil, false
					}
					lines = append(lines, ln)
				}
			}
			for j, ln := range lines {
				rows = append(rows, flatSidebarRow{infoIdx: i, right: ln, cont: j > 0})
			}
		}
	}
	return rows, true
}

func sidebarFitsWidth(termW, logoPadW, gapLen int, flat []flatSidebarRow) bool {
	if termW <= 0 {
		return true
	}
	for _, fr := range flat {
		if fr.blank {
			continue
		}
		if logoPadW+gapLen+runewidth.StringWidth(fr.right) > termW {
			return false
		}
	}
	return true
}

func (d *Display) printLogoSidebar(infos []modules.ModuleInfo, logo []string, logoPadW int, gap string, flat []flatSidebarRow, maxKeyWidth int, keyColor, titleColor string) {
	logoRow := 0

	emitLeft := func() string {
		var left string
		if logoRow < len(logo) {
			pad := logoPadW - runewidth.StringWidth(logo[logoRow])
			if pad > 0 {
				left = logo[logoRow] + strings.Repeat(" ", pad)
			} else {
				left = logo[logoRow]
			}
			logoRow++
		} else if len(logo) > 0 {
			left = strings.Repeat(" ", logoPadW)
		}
		return left
	}

	for _, fr := range flat {
		left := emitLeft()

		if fr.blank {
			if left != "" {
				fmt.Print(Paint(left, keyColor))
			}
			fmt.Println()
			continue
		}

		info := infos[fr.infoIdx]
		if left != "" {
			fmt.Print(Paint(left, keyColor))
			fmt.Print(gap)
		}

		if fr.right == "" {
			fmt.Println()
			continue
		}

		printSidebarRightPiece(info, fr.infoIdx, maxKeyWidth, fr.right, fr.cont, keyColor, titleColor)
		fmt.Println()
	}

	for logoRow < len(logo) {
		left := emitLeft()
		if left != "" {
			fmt.Print(Paint(left, keyColor))
		}
		fmt.Println()
	}
}

func printSidebarRightPiece(info modules.ModuleInfo, infoIdx, maxKeyWidth int, right string, continuation bool, keyColor, titleColor string) {
	switch {
	case info.Key == "separator":
		fmt.Print(Paint(right, keyColor))
		return
	case continuation && infoIdx == 0 && info.Value == "":
		fmt.Print(PaintTitle(right, titleColor))
	case continuation:
		if info.UsagePercent > 0 && info.Key != "" {
			fmt.Print(Paint(right, usageColor(info.UsagePercent)))
		} else {
			fmt.Print(right)
		}
	case infoIdx == 0 && info.Value == "" && info.Key != "":
		fmt.Print(PaintTitle(right, titleColor))
	default:
		isTitle := infoIdx == 0 && info.Value == ""
		coloredKey, rest := splitColored(right, maxKeyWidth, keyColor, isTitle)
		fmt.Print(coloredKey)
		if info.UsagePercent > 0 {
			fmt.Print(Paint(rest, usageColor(info.UsagePercent)))
		} else {
			fmt.Print(rest)
		}
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
	tw := getTerminalWidth()

	for i, info := range infos {
		if info.Key == "separator" {
			sepW := 40
			if tw > 0 {
				sepW = tw
			}
			sepLine := strings.Repeat("─", sepW)
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
		if tw > 0 && lineWidth > tw {
			truncateAt := tw - 3
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
	for _, fd := range []int{int(os.Stdout.Fd()), int(os.Stderr.Fd())} {
		if w, _, err := term.GetSize(fd); err == nil && w > 0 {
			return w
		}
	}
	if s := strings.TrimSpace(os.Getenv("COLUMNS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return 0
}

func wrapParagraph(s string, maxW int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if maxW <= 0 {
		return []string{s}
	}
	if runewidth.StringWidth(s) <= maxW {
		return []string{s}
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{s}
	}

	var lines []string
	var b strings.Builder
	w := 0
	flush := func() {
		if b.Len() == 0 {
			return
		}
		lines = append(lines, strings.TrimSpace(b.String()))
		b.Reset()
		w = 0
	}

	for _, word := range words {
		ww := runewidth.StringWidth(word)
		if ww > maxW {
			flush()
			for _, part := range splitLongWord(word, maxW) {
				lines = append(lines, part)
			}
			continue
		}
		add := ww
		if w > 0 {
			add++
		}
		if w+add <= maxW {
			if w > 0 {
				b.WriteByte(' ')
				w++
			}
			b.WriteString(word)
			w += ww
		} else {
			flush()
			b.WriteString(word)
			w = ww
		}
	}
	flush()
	return lines
}

func splitLongWord(word string, maxW int) []string {
	if maxW <= 0 {
		return []string{word}
	}
	rs := []rune(word)
	var out []string
	for len(rs) > 0 {
		var acc []rune
		cw := 0
		for len(rs) > 0 {
			r := rs[0]
			rw := runewidth.RuneWidth(r)
			if len(acc) == 0 && rw > maxW {
				acc = append(acc, r)
				rs = rs[1:]
				break
			}
			if cw+rw > maxW {
				break
			}
			acc = append(acc, r)
			cw += rw
			rs = rs[1:]
		}
		if len(acc) == 0 {
			break
		}
		out = append(out, string(acc))
	}
	if len(out) == 0 {
		return []string{word}
	}
	return out
}
