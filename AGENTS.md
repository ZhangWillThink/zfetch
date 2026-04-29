# AGENTS.md

## Build & Test

```bash
bash scripts/build.sh   # Build all platforms (Linux/macOS, amd64/arm64) to dist/

go build              # Linux native
GOOS=darwin GOARCH=arm64 go build ./...   # macOS Apple Silicon
GOOS=darwin GOARCH=amd64 go build ./...   # macOS Intel
GOOS=windows GOARCH=amd64 go build ./...  # Windows
```

No tests, no linter, no CI configured. Only `go build` for verification.

## Architecture

```
main.go            → CLI entrypoint, flag parsing, wire-up
config/            → JSONC config loader (custom comment-stripping parser)
display/           → Render engine: color ANSI, logo lookup, left/right column layout
modules/           → One file per info module, auto-registered via init()
internal/sysinfo/  → Platform-specific data collection (build tags)
presets/           → default.jsonc, all.jsonc
```

Zero dependencies beyond Go stdlib.

## Platform-specific code (`internal/sysinfo/`)

Three files with build tags:

| File         | Tag                  | Status   |
| ------------ | -------------------- | -------- |
| `linux.go`   | `//go:build linux`   | Complete |
| `darwin.go`  | `//go:build darwin`  | Complete |
| `windows.go` | `//go:build windows` | Partial  |

**Critical rule**: Every file in this package MUST have a build tag. If one file has no tag, it compiles into all platforms and causes "redeclared" errors. The linux file originally lacked a tag — don't remove it.

All functions return the same signatures defined in `sysinfo.go`. New platforms must implement every function.

## Adding a new module

1. Add file to `modules/` with `init() { Register(&MyModule{}) }` and `Name()`/`Run()` methods
2. Add the module name string to `getAllModules()` in `main.go`
3. Add it to the default structure in `config.DefaultConfig()` (`config/config.go`)

The module registry uses a blank import `_ "github.com/WillZhang/zfetch/modules"` in `main.go` to trigger all `init()` functions.

## Config

- Format: JSONC (JSON with `//` and `/* */` comments)
- Default path: `~/.config/zfetch/config.jsonc`
- Custom parser in `config/config.go` — uses standard `encoding/json` after stripping comments

## Logos (`display/logo.go`)

`detectOSLogo()` first checks if `sysinfo.GetOS().ID` matches a logo key, then tries fuzzy name match, then checks OS-specific release files, and finally falls back to `"linux"`. Currently only Linux distro logos exist. macOS and Windows will fall through to the `"default"` logo.
