# zfetch

A fast, feature-rich system information tool for the terminal — inspired by neofetch / screenfetch, written in Go.

[中文文档](./README-ZH-CN.md)

- **Zero dependencies** — pure Go standard library
- **Cross-platform** — Linux, macOS, Windows
- **Customizable** — JSONC config, presets, colors, custom logo
- **Script-friendly** — pipe mode with plain-text output

## Quick Start

### Install via curl (Linux & macOS)

```bash
curl -fsSL https://github.com/ZhangWillThink/zfetch/releases/latest/download/install.sh | bash
```

Or specify a version:

```bash
curl -fsSL https://github.com/ZhangWillThink/zfetch/releases/latest/download/install.sh | ZFETCH_VERSION=v0.2.0 bash
```

### Build from source

```bash
go build
./zfetch
```

## Features

| Module     | Info                | Linux | macOS | Windows |
| ---------- | ------------------- | :---: | :---: | :-----: |
| title      | User & Host         |   ✓   |   ✓   |    —    |
| os         | OS Name & Version   |   ✓   |   ✓   |    ✓    |
| host       | Host Machine        |   ✓   |   ✓   |    —    |
| kernel     | Kernel Details      |   ✓   |   ✓   |    ✓    |
| uptime     | System Uptime       |   ✓   |   ✓   |    —    |
| packages   | Package Count       |   ✓   |   ✓   |    —    |
| shell      | Shell & Version     |   ✓   |   ✓   |    ✓    |
| resolution | Display Resolution  |   ✓   |   ✓   |    —    |
| de         | Desktop Environment |   ✓   |   ✓   |    —    |
| wm         | Window Manager      |   ✓   |   ✓   |    —    |
| terminal   | Terminal Emulator   |   ✓   |   ✓   |    ✓    |
| cpu        | CPU Model & Cores   |   ✓   |   ✓   |    ✓    |
| gpu        | GPU Info (multi)    |   ✓   |   ✓   |    ✓    |
| memory     | RAM Usage           |   ✓   |   ✓   |    —    |
| swap       | Swap Usage          |   ✓   |   ✓   |    —    |
| disk       | Disk Usage (multi)  |   ✓   |   ✓   |    ✓    |
| battery    | Battery Status      |   ✓   |   ✓   |    —    |
| localip    | Local IP Address    |   ✓   |   ✓   |    —    |
| locale     | System Locale       |   ✓   |   ✓   |    —    |

## Usage

```bash
# Basic
zfetch

# Upgrade to latest version
zfetch upgrade

# Uninstall
zfetch uninstall

# Load a preset
zfetch -c default

# Custom module order
zfetch -s "os:kernel:uptime:shell:cpu:memory:disk"

# Set logo and color
zfetch --logo ubuntu --color cyan

# Pipe mode (plain text, no colors)
zfetch --pipe
```

### Options

| Flag                      | Description                              |
| ------------------------- | ---------------------------------------- |
| `-h`, `--help`            | Show help                                |
| `-v`, `--version`         | Show version                             |
| `-c`, `--config <file>`   | Load config file                         |
| `-s`, `--structure <str>` | Set module display order (`:` separated) |
| `--logo <name>`           | Set logo                                 |
| `--color <name>`          | Set accent color for keys & title        |
| `--color-keys <name>`     | Set color for keys only                  |
| `--pipe`                  | Disable colors and logo                  |
| `--stat`                  | Show per-module timing                   |
| `--list-modules`          | List all modules                         |
| `--list-logos`            | List available logos                     |
| `--list-presets`          | List available presets                   |
| `--gen-config`            | Show default config path                 |
| `upgrade`                 | Upgrade to the latest version            |
| `uninstall`               | Uninstall zfetch                         |

### Colors

`black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`,
`bright_black`, `bright_red`, `bright_green`, `bright_yellow`,
`bright_blue`, `bright_magenta`, `bright_cyan`, `bright_white`

## Configuration

Config files use **JSONC** format (JSON with `//` and `/* */` comments).

**Default path:** `~/.config/zfetch/config.jsonc`

```jsonc
{
  "structure": "title:separator:os:host:kernel:uptime:packages:shell:resolution:de:wm:terminal:cpu:gpu:memory:swap:disk:battery:localip:locale",
  "separator": "~",
  "colorKeys": "cyan",
  "colorTitle": "bright_cyan",
  "pipe": false,
  "logo": ""
}
```

Use `zfetch --list-config-paths` to see all search paths.

## Building

```bash
bash scripts/build.sh   # Build all platforms to dist/

# Linux (native)
go build

# macOS Intel
GOOS=darwin GOARCH=amd64 go build ./...

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build ./...

# Windows
GOOS=windows GOARCH=amd64 go build ./...
```
