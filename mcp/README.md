# 🌐 Skillyfiy MCP Server (Self-Hostable on Vercel Free Plan)

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2FAgentNex%2Fskillyfiy&root-directory=mcp&project-name=skillyfiy-mcp&repository-name=skillyfiy-mcp)
[![License: MIT](https://img.shields.io/badge/License-MIT-emerald.svg)](../LICENSE)
[![Zero React](https://img.shields.io/badge/Framework-Zero%20React%20%2F%20Pure%20Serverless-06B6D4.svg)](#)
[![Vercel Free Plan](https://img.shields.io/badge/Vercel-100%25%20Free%20Hobby-10B981.svg)](#)

A high-performance, self-hostable **Model Context Protocol (MCP)** server for the [Skillyfiy](https://github.com/AgentNex/skillyfiy) Context Optimization Suite. Deployable to Vercel's free tier with a single click, enabling any AI agent to discover skills, read documentation, search MCP servers, and audit prompt context token overhead.

---

## ⚡ 1-Click Instant Deployment

You can host your own private, free instance in 30 seconds:

1. Click the **Deploy with Vercel** button above.
2. Sign in with GitHub or GitLab.
3. Vercel will clone the `/mcp` directory and deploy your endpoint automatically.
4. Your personal MCP server will be live at: `https://<your-project>.vercel.app/api/mcp`!

---

## 💡 Why Self-Host on Vercel?

* **100% Free Forever**: Runs within Vercel's generous Hobby tier (100,000 requests/day, 100 GB bandwidth).
* **Zero React / Zero Bloat**: Built purely with Web Standard serverless handlers (`mcp-handler`) and static HTML5/Tailwind. Sub-millisecond execution with no virtual DOM or hydration overhead.
* **Universal Agent Compatibility**: Works over standard Server-Sent Events (SSE) and HTTP stream with Claude Desktop, Cursor, Antigravity, Claude Code, Windsurf, Cline, and Roo Code.
* **Privacy & Isolation**: You own your endpoint and your server logs.

---

## 🛠️ Available MCP Tools

Once connected, your AI agent gains access to the following 6 tools:

| Tool | Parameters | Description |
|---|---|---|
| `search_skills` | `query?`, `category?`, `limit?` | Fuzzy search curated agent skills by keyword, category, or tag with token weights. |
| `get_skill_doc` | `skill_name` | Fetch complete `SKILL.md` rules, instructions, system prompts, and code examples. |
| `search_mcp_servers` | `query?`, `category?`, `limit?` | Discover MCP servers (database, devtools, cloud, memory) with their tool capabilities. |
| `get_mcp_config` | `server_name`, `client_type?` | Generates copy-paste configuration snippets for Claude, Cursor, Antigravity, etc. |
| `calculate_overhead` | `skills?`, `mcp_servers?` | Calculates total prompt token overhead and assesses context degradation risks. |
| `generate_agent_setup` | `client_type`, `endpoint_url` | Generates the exact client config snippet to connect your agent to this MCP server. |

---

## 🤖 Method 1: Automated Agent Setup (Recommended)

Don't want to edit configuration files manually? Let your AI agent configure itself!

1. Open [`AGENT_SETUP_PROMPT.md`](./AGENT_SETUP_PROMPT.md).
2. Copy the prompt block.
3. Paste it directly into your AI Agent (Cursor, Claude, Antigravity, Windsurf).
4. The agent will detect its environment, locate your config file, merge the server entry, and verify connectivity!

---

## 🔧 Method 2: Manual Client Configuration

Add your deployed Vercel endpoint (`https://<YOUR-PROJECT>.vercel.app/api/mcp`) to your AI client:

### 1. Claude Desktop
Edit your `claude_desktop_config.json`:
- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "skillyfiy": {
      "url": "https://<YOUR-PROJECT>.vercel.app/api/mcp"
    }
  }
}
```

### 2. Cursor IDE
Add to `.cursor/mcp.json` or in **Cursor Settings $\to$ MCP $\to$ Add New MCP Server**:
- **Name:** `skillyfiy`
- **Type:** `sse`
- **URL:** `https://<YOUR-PROJECT>.vercel.app/api/mcp`

### 3. Antigravity CLI / Gemini
Add to `~/.gemini/antigravity-cli/mcp_config.json`:
```json
{
  "skillyfiy": {
    "url": "https://<YOUR-PROJECT>.vercel.app/api/mcp"
  }
}
```

### 4. Claude Code CLI
Run directly in your terminal:
```bash
claude mcp add --transport sse skillyfiy https://<YOUR-PROJECT>.vercel.app/api/mcp
```

### 5. Windsurf
Add to `~/.codeium/windsurf/mcp_config.json`:
```json
{
  "mcpServers": {
    "skillyfiy": {
      "serverUrl": "https://<YOUR-PROJECT>.vercel.app/api/mcp"
    }
  }
}
```

---

## 💻 Local Development & Testing

You can also run the server locally on your machine:

```bash
cd mcp

# 1. Install dependencies
npm install

# 2. Run local test suite
npm test

# 3. Start local server
npm run dev
```

Visit `http://localhost:3000` in your browser to view the interactive dashboard, or test the health endpoint:
```bash
curl -s http://localhost:3000/api/health
```

---

## 📄 License

MIT © [AgentNex](https://github.com/AgentNex)
