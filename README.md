# zfetch

A fast, feature-rich system information tool for the terminal — inspired by neofetch / screenfetch, written in Go.

[中文文档](./README-ZH-CN.md)

- **Zero dependencies** — pure Go standard library
- **Cross-platform** — Linux, macOS, Windows
- **Customizable** — JSONC config, presets, colors, custom logo
- **Script-friendly** — pipe mode with plain-text output

## Quick Start

```bash
go build
./zfetch
```

## Features

| Module     | Info                | Linux | macOS | Windows |
| ---------- | ------------------- | :---: | :---: | :-----: |
| title      | User & Host         |   ✓   |   ✓   |    —    |
| os         | OS Name & Version   |   ✓   |   ✓   |    ✓    |
| kernel     | Kernel Details      |   ✓   |   ✓   |    ✓    |
| uptime     | System Uptime       |   ✓   |   ✓   |    —    |
| packages   | Package Count       |   ✓   |   ✓   |    —    |
| shell      | Shell & Version     |   ✓   |   ✓   |    ✓    |
| resolution | Display Resolution  |   ✓   |   ✓   |    —    |
| de         | Desktop Environment |   ✓   |   ✓   |    —    |
| wm         | Window Manager      |   ✓   |   ✓   |    —    |
| terminal   | Terminal Emulator   |   ✓   |   ✓   |    ✓    |
| cpu        | CPU Model & Cores   |   ✓   |   ✓   |    ✓    |
| gpu        | GPU Name            |   ✓   |   ✓   |    ✓    |
| memory     | RAM Usage           |   ✓   |   ✓   |    —    |
| disk       | Disk Usage          |   ✓   |   ✓   |    —    |

## Usage

```bash
# Basic
zfetch

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

### Colors

`black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`,
`bright_black`, `bright_red`, `bright_green`, `bright_yellow`,
`bright_blue`, `bright_magenta`, `bright_cyan`, `bright_white`

## Configuration

Config files use **JSONC** format (JSON with `//` and `/* */` comments).

**Default path:** `~/.config/zfetch/config.jsonc`

```jsonc
{
  "structure": "title:separator:os:kernel:uptime:packages:shell:resolution:de:wm:terminal:cpu:gpu:memory:disk",
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
# Linux (native)
go build

# macOS Intel
GOOS=darwin GOARCH=amd64 go build ./...

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build ./...

# Windows
GOOS=windows GOARCH=amd64 go build ./...
```
