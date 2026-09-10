<div align="center">

# ⚡ Skillyfiy CLI
**Surgically search, preview, multi-select, and purge unused AI agent skills & MCP servers.**

[![Go Version](https://img.shields.io/github/go-mod/go-version/AgentNex/skillyfiy?style=flat-square&color=10B981)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/AgentNex/skillyfiy?style=flat-square&color=38BDF8)](https://github.com/AgentNex/skillyfiy/releases)
[![License](https://img.shields.io/badge/License-MIT-slate?style=flat-square)](LICENSE)
[![Zero-CGO](https://img.shields.io/badge/CGO-Zero%20Dependencies-059669?style=flat-square)](#)

</div>

---

### 💡 Why Skillyfiy?
Accumulating hundreds of agent skills and MCP servers pollutes context windows, degrades reasoning performance, and inflates system prompt token costs. **Skillyfiy** provides an interactive, split-pane Terminal User Interface (TUI) designed to browse, inspect, and bulk-prune agent overhead with sub-millisecond responsiveness.

```
┌── SKILLYFIY v0.1.0 ────────────────────────────┬───────────────────────────────┐
│ / Search: [ scraper                          ] │ INSPECTOR                     │
├────────────────────────────────────────────────┼───────────────────────────────┤
│ ▶ [✓] [SKILL] web_scraper.py           1.2 KB  │ Target: web_scraper.py        │
│   [ ] [MCP]   mcp-server-playwright   14.8 KB  │ Type:   Agent Skill (Python)  │
│   [✓] [SKILL] doc_indexer.md           0.8 KB  │ System Tokens: ~840 tokens    │
├────────────────────────────────────────────────┴───────────────────────────────┤
│ [2/1,542 Selected]  •  ~2,040 Tokens Reclaimed                                 │
│ [Space] Toggle • [s] Span Range • [Ctrl+A] All • [Enter] Purge • [Ctrl+Q] Quit │
└───────────────────────────────────────────────────────────────────────────────┘
```

---

## ⚡ Quick Install

### Method 1: Universal 1-Line Install (Termux / Linux / macOS)
```bash
curl -fsSL https://raw.githubusercontent.com/AgentNex/skillyfiy/main/install.sh | bash
```

### Method 2: Go Toolchain
```bash
go install github.com/AgentNex/skillyfiy@latest
```

### 🤖 Install via AI Agent

<details>
<summary>📋 <b>Click to reveal prompt for installing Skillyfiy using your AI agent</b></summary>

Hover over the block below and click the top-right copy icon to copy the prompt:

```
Install the Skillyfiy CLI tool on my system:

1. Check if Go is installed (`go version`) or use curl to run the installer:
   curl -fsSL https://raw.githubusercontent.com/AgentNex/skillyfiy/main/install.sh | bash
2. If Go is available, compile from source:
   go install github.com/AgentNex/skillyfiy@latest
3. Verify the installation by running:
   skillyfiy -v
4. Provide a quick summary of the available flags and how to start the tool.
```

</details>

## 🎮 Navigation & Keybindings

| Key | Action |
|---|---|
| `Space` / `x` | Toggle single item |
| `s` | Toggle Skillyfiy Span (Visual Range mode) |
| `Ctrl+A` | Toggle select all visible items |
| `Tab` | Switch between Search and List pane |
| `Enter` | Review and confirm purge |
| `Ctrl+Q` / `Ctrl+C` | Exit without making changes |

## ⚙️ CLI Flags

```bash
# Launch interactive TUI
skillyfiy

# Print version information
skillyfiy -v
skillyfiy --version

# Specify custom skills directory
skillyfiy --skills /path/to/custom/skills

# Specify custom MCP configuration file
skillyfiy --mcp /path/to/claude_desktop_config.json

# Display help menu
skillyfiy --help
```

## 🗑️ Uninstallation

Because Skillyfiy is a single, zero-dependency static binary that leaves no background daemons or hidden system directories, uninstalling it is instant:

### Termux (Android):
```bash
rm -f $PREFIX/bin/skillyfiy
```

### Linux / macOS:
```bash
rm -f /usr/local/bin/skillyfiy ~/.local/bin/skillyfiy $(go env GOPATH 2>/dev/null)/bin/skillyfiy
```

### Optional (Remove Cloned Source):
```bash
rm -rf ~/skillyfiy
```

## 📄 License

Released under the [MIT License](LICENSE).
