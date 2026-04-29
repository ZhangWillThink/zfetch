# zfetch

一个快速、功能丰富的终端系统信息工具——受 neofetch / screenfetch 启发，用 Go 编写。

[English](./README.md)

- **零依赖** — 纯 Go 标准库
- **跨平台** — Linux、macOS、Windows
- **可定制** — JSONC 配置、预设、颜色、自定义 Logo
- **脚本友好** — 管道模式输出纯文本

## 快速开始

### 通过 curl 安装（Linux & macOS）

```bash
curl -fsSL https://github.com/ZhangWillThink/zfetch/releases/latest/download/install.sh | bash
```

或指定版本：

```bash
ZFETCH_VERSION=v0.1.1 curl -fsSL https://github.com/ZhangWillThink/zfetch/releases/download/v0.1.1/install.sh | bash
```

### 从源码构建

```bash
go build
./zfetch
```

## 功能

| 模块       | 信息             | Linux | macOS | Windows |
| ---------- | ---------------- | :---: | :---: | :-----: |
| title      | 用户和主机名     |   ✓   |   ✓   |    —    |
| os         | 系统名称和版本   |   ✓   |   ✓   |    ✓    |
| kernel     | 内核详情         |   ✓   |   ✓   |    ✓    |
| uptime     | 系统运行时间     |   ✓   |   ✓   |    —    |
| packages   | 软件包数量       |   ✓   |   ✓   |    —    |
| shell      | Shell 及版本     |   ✓   |   ✓   |    ✓    |
| resolution | 显示器分辨率     |   ✓   |   ✓   |    —    |
| de         | 桌面环境         |   ✓   |   ✓   |    —    |
| wm         | 窗口管理器       |   ✓   |   ✓   |    —    |
| terminal   | 终端模拟器       |   ✓   |   ✓   |    ✓    |
| cpu        | CPU 型号和核心数 |   ✓   |   ✓   |    ✓    |
| gpu        | 显卡名称         |   ✓   |   ✓   |    ✓    |
| memory     | 内存使用情况     |   ✓   |   ✓   |    —    |
| disk       | 磁盘使用情况     |   ✓   |   ✓   |    —    |

## 用法

```bash
# 基本使用
zfetch

# 使用预设配置
zfetch -c default

# 自定义模块顺序
zfetch -s "os:kernel:uptime:shell:cpu:memory:disk"

# 设置 Logo 和颜色
zfetch --logo ubuntu --color cyan

# 管道模式（纯文本，无颜色）
zfetch --pipe
```

### 选项

| 选项                       | 说明                         |
| -------------------------- | ---------------------------- |
| `-h`, `--help`             | 显示帮助                     |
| `-v`, `--version`          | 显示版本                     |
| `-c`, `--config <文件>`    | 加载配置文件                 |
| `-s`, `--structure <结构>` | 设置模块显示顺序（`:` 分隔） |
| `--logo <名称>`            | 设置 Logo                    |
| `--color <名称>`           | 设置主色（键和标题）         |
| `--color-keys <名称>`      | 仅设置键的颜色               |
| `--pipe`                   | 禁用颜色和 Logo              |
| `--stat`                   | 显示各模块耗时               |
| `--list-modules`           | 列出所有模块                 |
| `--list-logos`             | 列出可用 Logo                |
| `--list-presets`           | 列出可用预设                 |
| `--gen-config`             | 显示默认配置路径             |

### 颜色

`black`、`red`、`green`、`yellow`、`blue`、`magenta`、`cyan`、`white`、
`bright_black`、`bright_red`、`bright_green`、`bright_yellow`、
`bright_blue`、`bright_magenta`、`bright_cyan`、`bright_white`

## 配置

配置文件使用 **JSONC** 格式（支持 `//` 和 `/* */` 注释）。

**默认路径:** `~/.config/zfetch/config.jsonc`

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

使用 `zfetch --list-config-paths` 查看所有搜索路径。

## 构建

```bash
# Linux (原生)
go build

# macOS Intel
GOOS=darwin GOARCH=amd64 go build ./...

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build ./...

# Windows
GOOS=windows GOARCH=amd64 go build ./...
```
