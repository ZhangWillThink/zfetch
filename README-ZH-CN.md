# zfetch

一个快速、功能丰富的终端系统信息工具——受 neofetch / screenfetch 启发，用 Go 编写。

[English](./README.md)

- **单个静态二进制** — 体积小，目标机无需额外安装运行时
- **跨平台** — Linux、macOS、Windows
- **可定制** — JSONC 配置、预设、颜色、自定义 Logo
- **脚本友好** — 管道模式输出纯文本

## 快速开始

### 通过 curl 安装（Linux & macOS）

```bash
curl -fsSL https://github.com/ZhangWillThink/zfetch/releases/latest/download/install.sh | bash
```

或固定在某个发布版本（可复现安装）：

```bash
curl -fsSL https://github.com/ZhangWillThink/zfetch/releases/latest/download/install.sh | ZFETCH_VERSION=v0.5.1 bash
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
| host       | 主机型号         |   ✓   |   ✓   |    —    |
| kernel     | 内核详情         |   ✓   |   ✓   |    ✓    |
| uptime     | 系统运行时间     |   ✓   |   ✓   |    —    |
| packages   | 软件包数量       |   ✓   |   ✓   |    —    |
| shell      | Shell 及版本     |   ✓   |   ✓   |    ✓    |
| resolution | 显示器分辨率     |   ✓   |   ✓   |    —    |
| de         | 桌面环境         |   ✓   |   ✓   |    —    |
| wm         | 窗口管理器       |   ✓   |   ✓   |    —    |
| terminal   | 终端模拟器       |   ✓   |   ✓   |    ✓    |
| cpu        | CPU 型号和核心数 |   ✓   |   ✓   |    ✓    |
| gpu        | 显卡信息（多卡） |   ✓   |   ✓   |    ✓    |
| memory     | 内存使用情况     |   ✓   |   ✓   |    —    |
| swap       | 交换分区使用     |   ✓   |   ✓   |    —    |
| disk       | 磁盘使用（多盘） |   ✓   |   ✓   |    ✓    |
| battery    | 电池状态         |   ✓   |   ✓   |    —    |
| localip    | 本地 IP 地址     |   ✓   |   ✓   |    —    |
| locale     | 系统语言区域     |   ✓   |   ✓   |    —    |

## 用法

```bash
# 基本使用
zfetch

# 升级到最新版本
zfetch upgrade

# 卸载
zfetch uninstall

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
| `--gen-config`            | 将默认配置写入 ~/.config/zfetch/config.jsonc |
| `upgrade`                  | 升级到最新版本               |
| `uninstall`                | 卸载 zfetch                  |

### 颜色

`black`、`red`、`green`、`yellow`、`blue`、`magenta`、`cyan`、`white`、
`bright_black`、`bright_red`、`bright_green`、`bright_yellow`、
`bright_blue`、`bright_magenta`、`bright_cyan`、`bright_white`

## 配置

配置文件使用 **JSONC** 格式（支持 `//` 和 `/* */` 注释）。

**默认路径:** `~/.config/zfetch/config.jsonc`。若文件存在且在未传入 `-c` / `--config` 时运行 `zfetch`，会自动加载；需要预设或绝对路径时用 `-c` 指定。

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

使用 `zfetch --list-config-paths` 查看所有搜索路径。

## 构建

```bash
bash scripts/build.sh    # 输出 dist 并用 -ldflags 写入版本（环境变量 ZFETCH_VERSION，否则为当前 git 标签，否则 v0.0.0-dev+<短哈希>）

# Linux（原生）
go build ./...           # 默认内置版本为 v0.0.0-dev，除非自行传入与 build.sh 相同的 -ldflags

# macOS Intel
GOOS=darwin GOARCH=amd64 go build ./...

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build ./...

# Windows
GOOS=windows GOARCH=amd64 go build ./...
```
