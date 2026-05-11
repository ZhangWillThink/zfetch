package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/WillZhang/zfetch/config"
	"github.com/WillZhang/zfetch/display"
	"github.com/WillZhang/zfetch/internal/uninstall"
	"github.com/WillZhang/zfetch/internal/upgrade"
	"github.com/WillZhang/zfetch/modules"
	_ "github.com/WillZhang/zfetch/modules"
)

func main() {
	var (
		showHelp        bool
		showVersion     bool
		listModules     bool
		listPresets     bool
		listLogos       bool
		listConfigPaths bool
		printStructure  bool
		genConfig       bool
		statMode        bool
		structStr       string
		configFile      string
		logoName        string
		colorKeys       string
		colorName       string
		piped           bool
	)

	flag.BoolVar(&showHelp, "help", false, "Display help")
	flag.BoolVar(&showHelp, "h", false, "Display help (shorthand)")
	flag.BoolVar(&showVersion, "version", false, "Display version")
	flag.BoolVar(&showVersion, "v", false, "Display version (shorthand)")
	flag.BoolVar(&listModules, "list-modules", false, "List all available modules")
	flag.BoolVar(&listPresets, "list-presets", false, "List available presets")
	flag.BoolVar(&listLogos, "list-logos", false, "List available logos")
	flag.BoolVar(&listConfigPaths, "list-config-paths", false, "List search paths for config files")
	flag.BoolVar(&printStructure, "print-structure", false, "Print the default structure")
	flag.BoolVar(&genConfig, "gen-config", false, "Generate default config file")
	flag.BoolVar(&statMode, "stat", false, "Show time usage for individual modules")
	flag.StringVar(&structStr, "structure", "", "Set structure of the fetch")
	flag.StringVar(&structStr, "s", "", "Set structure of the fetch (shorthand)")
	flag.StringVar(&configFile, "config", "", "Load a config file")
	flag.StringVar(&configFile, "c", "", "Load a config file (shorthand)")
	flag.StringVar(&logoName, "logo", "", "Set the logo to use")
	flag.StringVar(&colorKeys, "color-keys", "", "Set the color of keys")
	flag.StringVar(&colorName, "color", "", "Set the color for keys and title")
	flag.BoolVar(&piped, "pipe", false, "Disable colors and logo")

	flag.Parse()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		os.Exit(1)
	}()

	if flag.Arg(0) == "upgrade" {
		if err := upgrade.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if flag.Arg(0) == "uninstall" {
		if err := uninstall.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if showHelp {
		printHelp()
		return
	}

	if showVersion {
		fmt.Printf("zfetch %s (%s/%s)\n", upgrade.CurrentVersion, runtime.GOOS, runtime.GOARCH)
		return
	}

	if listModules {
		names := modules.AllModules()
		sort.Strings(names)
		fmt.Println("Available modules:")
		for _, m := range names {
			fmt.Printf("  %s\n", m)
		}
		return
	}

	if listPresets {
		fmt.Println("Available presets:")
		for _, p := range config.ListPresetNames() {
			fmt.Printf("  %s\n", p)
		}
		return
	}

	if listLogos {
		logos := display.ListLogos()
		sort.Strings(logos)
		fmt.Println("Available logos:")
		for _, l := range logos {
			fmt.Printf("  %s\n", l)
		}
		return
	}

	if listConfigPaths {
		fmt.Println("Config search paths:")
		for _, p := range config.ListConfigPaths() {
			fmt.Printf("  %s\n", p)
		}
		return
	}

	if printStructure {
		fmt.Println(config.DefaultConfig().Structure)
		return
	}

	if genConfig {
		cfgPath := config.FindDefaultConfig()
		if cfgPath == "" {
			fmt.Fprintf(os.Stderr, "Error: cannot determine home directory\n")
			os.Exit(1)
		}
		if _, err := os.Stat(cfgPath); err == nil {
			fmt.Fprintf(os.Stderr, "Config already exists at %s\n", cfgPath)
			os.Exit(1)
		}
		dir := filepath.Dir(cfgPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(cfgPath, []byte(defaultConfigContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config generated at: %s\n", cfgPath)
		return
	}

	cfg := config.DefaultConfig()

	if configFile != "" {
		presetPath, err := config.FindPreset(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		loadedCfg, err := config.LoadFromFile(presetPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		cfg = loadedCfg
	} else if dc := config.FindDefaultConfig(); dc != "" {
		if fi, err := os.Stat(dc); err == nil && !fi.IsDir() {
			loadedCfg, err := config.LoadFromFile(dc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not load %s: %v (using defaults)\n", dc, err)
			} else {
				cfg = loadedCfg
			}
		}
	}

	if structStr != "" {
		cfg.Structure = structStr
	}

	if logoName != "" {
		cfg.Logo = logoName
	}

	if colorKeys != "" {
		cfg.ColorKeys = colorKeys
	}
	if colorName != "" {
		cfg.ColorKeys = colorName
		cfg.ColorTitle = colorName
	}

	if piped {
		cfg.Pipe = true
	}

	if statMode {
		cfg.Pipe = true
		printModuleTimings(cfg.Structure)
		return
	}

	pipedOutput := cfg.Pipe || !term.IsTerminal(int(os.Stdout.Fd()))

	d := display.New(cfg, pipedOutput)
	d.Render()
}

func printModuleTimings(structure string) {
	for _, key := range strings.Split(structure, ":") {
		key = strings.TrimSpace(key)
		if key == "" || key == "separator" {
			continue
		}
		m := modules.Get(key)
		if m == nil {
			continue
		}
		start := time.Now()
		_ = m.Run()
		fmt.Printf("%s\t%s\n", key, time.Since(start).Round(time.Microsecond))
	}
}

const defaultConfigContent = `{
  // zfetch configuration file (JSONC format - JSON with comments)
  // Available modules: title, separator, os, kernel, uptime, packages,
  //   shell, resolution, de, wm, terminal, cpu, gpu, memory, swap,
  //   disk, host, battery, localip, locale
  "structure": "title:separator:os:kernel:uptime:packages:shell:resolution:de:wm:terminal:cpu:gpu:memory:swap:disk:host:battery:localip:locale",
  // "=" or "~" or ": " or " → "
  "separator": ": ",
  // Available colors: black, red, green, yellow, blue, magenta, cyan, white,
  //   bright_black, bright_red, bright_green, bright_yellow,
  //   bright_blue, bright_magenta, bright_cyan, bright_white
  "colorKeys": "default",
  "colorTitle": "default"
}
`

func printHelp() {
	fmt.Print(`zfetch - A fast and feature-rich system information tool

Usage: zfetch [options]
       zfetch upgrade        Upgrade to the latest version
       zfetch uninstall      Uninstall zfetch

Options:
  -h, --help              Display this help message
  -v, --version           Display version information
  --list-modules          List all available modules
  --list-presets          List available presets
  --list-logos            List available logos
  --list-config-paths     List search paths for config files
  --print-structure       Print the default structure
  --gen-config            Write default config to ~/.config/zfetch/config.jsonc
  --stat                  Show time usage for individual modules
  -s, --structure <str>   Set structure of the fetch
  -c, --config <file>     Load a config file
  --logo <name>           Set the logo to use
  --color-keys <color>    Set the color of keys
  --color <color>         Set the color for keys and title
  --pipe                  Disable colors and logo

Colors:
  black, red, green, yellow, blue, magenta, cyan, white,
  bright_black, bright_red, bright_green, bright_yellow,
  bright_blue, bright_magenta, bright_cyan, bright_white

Environment:
  NO_COLOR                  If set (any non-empty value), disable ANSI colors when using a TTY
  FORCE_COLOR               If 1/true/yes, force colors on (overrides NO_COLOR) when not piping

Structure example:
  zfetch -s "title:separator:os:kernel:uptime:shell:cpu:memory:disk"

Config files use JSONC format (JSON with comments).
When no -c/--config flag is passed, ~/.config/zfetch/config.jsonc is loaded if it exists.
Generate a default config file with: zfetch --gen-config
`)
}
