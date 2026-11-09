<div align="center">

# ⚡ Skillyfiy CLI
**Surgically search, preview, multi-select, and purge unused AI agent skills & MCP servers.**

[![Go Version](https://img.shields.io/github/go-mod/go-version/AgentNex/skillyfiy?style=flat-square&color=10B981)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/AgentNex/skillyfiy?style=flat-square&color=38BDF8)](https://github.com/AgentNex/skillyfiy/releases)
[![Deploy with Vercel](https://img.shields.io/badge/Deploy%20to-Vercel-black?style=flat-square&logo=vercel)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2FAgentNex%2Fskillyfiy&root-directory=mcp&project-name=skillyfiy-mcp&repository-name=skillyfiy-mcp)
[![License](https://img.shields.io/badge/License-MIT-slate?style=flat-square)](LICENSE)
[![Zero-CGO](https://img.shields.io/badge/CGO-Zero%20Dependencies-059669?style=flat-square)](#)

</div>

---

### 💡 Why Skillyfiy?
Accumulating hundreds of agent skills and MCP servers pollutes context windows, degrades reasoning performance, and inflates system prompt token costs. **Skillyfiy** provides an interactive, split-pane Terminal User Interface (TUI) designed to browse, inspect, and bulk-prune agent overhead with sub-millisecond responsiveness.

```
┌── SKILLYFIY [sky] v0.4.0 ──────────────────────┬───────────────────────────────┐
│ / Search: [ scraper           ] [Sort] [Filter]│ INSPECTOR                     │
├────────────────────────────────────────────────┼───────────────────────────────┤
│ ▶ [✓] [SKILL] web_scraper.py           1.2 KB  │ Target: web_scraper.py        │
│   [ ] [MCP]   mcp-server-playwright   14.8 KB  │ Type:   Agent Skill (Python)  │
│   [✓] [SKILL] doc_indexer.md           0.8 KB  │ System Tokens: ~840 tokens    │
├────────────────────────────────────────────────┴───────────────────────────────┤
│ [2/1,542 Selected]  •  ~2,040 Tokens Reclaimed                                 │
│ [i] Help Modal • [/] Search • [v] Multi-Select • [Enter] Purge                 │
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
<summary>📋 <b>Click to reveal prompt for installing Skillyfiy (sky) using your AI agent</b></summary>

Hover over the block below and click the top-right copy icon to copy the prompt:

```
Install the Skillyfiy (sky) CLI tool on my system:

1. Check if Go is installed (`go version`) or use curl to run the installer:
   curl -fsSL https://raw.githubusercontent.com/AgentNex/skillyfiy/main/install.sh | bash
2. If Go is available, compile from source:
   go install github.com/AgentNex/skillyfiy@latest
3. Verify the installation by running:
   sky -v
4. Provide a quick summary of the available flags and how to start the tool with 'sky'.
```

</details>

## 🔄 Updating to the Latest Version

If you have already installed an older version of Skillyfiy (`sky`), you can update to the latest release at any time:

### Method 1: Re-run 1-Line Installer (Recommended)
Simply re-run the universal installer. It will automatically detect your OS and architecture, download the latest release binary from GitHub, and safely replace your existing executable:
```bash
curl -fsSL https://raw.githubusercontent.com/AgentNex/skillyfiy/main/install.sh | bash
```

### Method 2: Go Toolchain
If you installed using Go, fetch and compile the latest release:
```bash
go install github.com/AgentNex/skillyfiy@latest
```

### Verify Your Version
Verify that your installation is up to date:
```bash
sky -v
# or
skillyfiy -v
```

## 🌈 Gradient Animation

Skillyfiy's banner sweeps on its own — no key press required. The 9-color palette is
interpolated into a continuous ramp and sampled **once per display column**, so the gradient
flows smoothly across the letters instead of stepping from one tinted block to the next.

| Surface | Behaviour |
|---|---|
| `SKILLYFIY` banner (full + compact) | Sweeps continuously |
| `SKILL` / `MCP` item badges | Sweep, trailing the banner by 10% of the loop |
| Reclaimed-token footer accent | Sweeps, trailing by 20% of the loop |
| Help & confirm modal titles | Gradient-colored, frozen while the modal is open |
| Sort/filter pills, search box | Static |

The clock runs at a fixed 20 FPS (50 ms per frame) and **drops frames** rather than queuing
work, so a slow or mobile terminal simply gets fewer frames instead of falling behind.
Opening the Help overlay or the purge-confirmation modal pauses the sweep; closing it resumes
from exactly the phase it stopped on.

### Modes

Cycle modes with the `/anim` command (see [Command Mode](#-command-mode)):

```
OFF  ->  SLOW  ->  NORMAL  ->  FAST  ->  HUE  ->  OFF
```

The tool launches in **NORMAL**, so the first `/anim` moves to `FAST`.

| Mode | Loop duration | Notes |
|---|---|---|
| `OFF` | — | No animation; redraws stop entirely |
| `SLOW` | 12s | |
| `NORMAL` | 6s | Launch default |
| `FAST` | 3s | |
| `HUE` | 6s sweep | Also rotates the whole palette's hue, a full turn every 20s |

### Opting out

Every animation mode is session-only, so nothing is written to disk. Disable it for a run with:

```bash
sky --no-animation
```

The animation also stays off automatically when `NO_COLOR` is set or `TERM=dumb`, since a
color-less terminal cannot show a sweep.

## ⌨️ Command Mode

Press `Ctrl+K` to turn the search box into a dedicated command line. A leading `/` is optional,
so `/anim` and `anim` both work. Press `Enter` to run the command or `Esc` to return to normal
filtering.

| Command | Effect |
|---|---|
| `/anim` | Cycle the gradient animation: OFF → SLOW → NORMAL → FAST → HUE → OFF |

The result is announced in the footer for a couple of seconds. While a command is being typed,
the list is **not** filtered, so command text never hides your items. An active multi-select or
pinned span still takes priority in the footer.

## 🎮 Navigation & Keybindings

| Key | Action |
|---|---|
| `↑` / `↓` / `k` / `j` | Move cursor through items / scroll inspector |
| `Space` / `x` | Toggle single item selection |
| `v` | Turn Multi-Select mode ON / OFF |
| `s` | Pin span anchor on start item; scroll; press `s` again on end item to select range |
| `Ctrl+A` | Toggle select/deselect all visible items |
| `/` | Focus search bar (**press `/` again to exit/blur search**) |
| `Ctrl+K` | Open **Command Mode** — type `/anim` to cycle the gradient animation |
| `Ctrl+S` / `F1` | Cycle 6 sort modes (Newest, Oldest, Largest, Smallest, A-Z, Z-A) |
| `Ctrl+T` / `F2` | Cycle 3 resource filters (All, Skills, MCP Servers) |
| `Tab` | Switch focus between List pane and Inspector pane |
| `Enter` | Review and confirm purge of selected items |
| `i` / `?` | **Toggle Help Modal overlay** showing all controls |
| `Ctrl+Q` / `Esc` | Exit sky |

## ⚙️ CLI Flags

Launch directly with `sky` (or `skillyfiy`):

```bash
# Launch interactive TUI
sky

# List all discovered skills and MCP servers non-interactively
skillyfiy list
sky list

# Print version information
sky -v
sky --version

# Specify custom skills directory
sky --skills /path/to/custom/skills

# Specify custom MCP configuration file
sky --mcp /path/to/claude_desktop_config.json

# Start with the gradient animation disabled (cycle it on later with /anim)
sky --no-animation

# Display help menu
sky --help
```

## 🌐 Self-Hostable Free MCP Server (Vercel)

Skillyfiy includes a production-ready, self-hostable **Model Context Protocol (MCP) Server** running on Vercel's free serverless tier with **zero React/Next.js overhead**. It allows AI coding assistants (Claude Desktop, Cursor, Antigravity, Claude Code, Windsurf, Cline) to query available agent skills, fetch full skill documentation, generate copy-paste MCP configurations, and dynamically evaluate prompt token overhead.

### 🚀 1-Click Free Deploy to Vercel

Deploy your own private MCP server to Vercel's free tier with a single click:

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2FAgentNex%2Fskillyfiy&root-directory=mcp&project-name=skillyfiy-mcp&repository-name=skillyfiy-mcp)

* **Endpoints**: Live SSE & HTTP POST at `/api/mcp`
* **Interactive Dashboard**: Modern dark-mode web dashboard served from `/` (built with vanilla HTML5/Tailwind/ES2022)
* **Dedicated MCP Documentation**: See [mcp/README.md](mcp/README.md) for complete architecture, tool specs, and client configs.
* **Autonomous Setup Prompt**: Copy [mcp/AGENT_SETUP_PROMPT.md](mcp/AGENT_SETUP_PROMPT.md) to have your AI agent automatically configure itself with your deployed endpoint.

## 🗑️ Uninstallation

Because Skillyfiy (`sky`) is a single, zero-dependency static binary that leaves no background daemons or hidden system directories, uninstalling it is instant:

### Termux (Android):
```bash
rm -f $PREFIX/bin/sky $PREFIX/bin/skillyfiy
```

### Linux / macOS:
```bash
rm -f /usr/local/bin/sky /usr/local/bin/skillyfiy ~/.local/bin/sky ~/.local/bin/skillyfiy $(go env GOPATH 2>/dev/null)/bin/skillyfiy
```

### Optional (Remove Cloned Source):
```bash
rm -rf ~/skillyfiy
```

## 📄 License

Released under the [MIT License](LICENSE).
