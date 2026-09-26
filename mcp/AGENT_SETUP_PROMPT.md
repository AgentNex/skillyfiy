# Master Prompt: Automatic AI Agent Self-Configuration for Skillyfiy MCP Server

> **Instructions for the User:**
> Copy the prompt block below and paste it directly into your AI Agent (Cursor, Claude Desktop, Antigravity, Claude Code, Windsurf, Cline, or Roo Code). The agent will automatically detect its environment, locate your configuration file, safely configure the Skillyfiy MCP server, and verify the connection!

---

```markdown
You are an autonomous AI Agent assistant. Your task is to automatically configure the **Skillyfiy Model Context Protocol (MCP) Server** into my active AI tool environment so that you can discover agent skills, read documentation, and optimize context token overhead.

### Target MCP Server Information
- **Server Name:** `skillyfiy`
- **Transport Type:** HTTP / Server-Sent Events (SSE)
- **Endpoint URL:** `https://<YOUR-VERCEL-DOMAIN>.vercel.app/api/mcp`
*(Note: If I have not provided my custom Vercel domain yet, please prompt me for my Vercel URL, or ask if I want to use a default endpoint).*

---

### Step-by-Step Execution Directives

1. **Detect Agent Environment & Config File:**
   Identify which AI Agent environment you are currently operating in, and locate the appropriate MCP configuration file on my system:
   - **Claude Desktop (macOS):** `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Claude Desktop (Windows):** `%APPDATA%/Claude/claude_desktop_config.json`
   - **Claude Desktop (Linux):** `~/.config/Claude/claude_desktop_config.json`
   - **Cursor IDE:** `.cursor/mcp.json` (or `~/.config/Cursor/User/globalStorage/cursor.mcp/mcp.json`)
   - **Antigravity CLI / Gemini:** `~/.gemini/antigravity-cli/mcp_config.json`
   - **Claude Code CLI:** Run `claude mcp add --transport sse skillyfiy <ENDPOINT_URL>`
   - **Windsurf IDE:** `~/.codeium/windsurf/mcp_config.json`
   - **Cline / Roo Code:** `cline_mcp_settings.json`

2. **Read & Safely Merge Configuration:**
   - If the configuration file exists, read its existing contents first.
   - Do NOT overwrite or delete any existing MCP servers in the file.
   - Add or merge the `skillyfiy` server entry into the `mcpServers` object:
     ```json
     {
       "mcpServers": {
         "skillyfiy": {
           "url": "https://<YOUR-VERCEL-DOMAIN>.vercel.app/api/mcp"
         }
       }
     }
     ```
   - (For Windsurf, use `"serverUrl": "https://<YOUR-VERCEL-DOMAIN>.vercel.app/api/mcp"`).
   - Write the cleanly formatted JSON back to disk.

3. **Verify Connectivity:**
   - Make a test GET request to `https://<YOUR-VERCEL-DOMAIN>.vercel.app/api/health` to confirm the endpoint is reachable.
   - Confirm that the server returns `status: "healthy"` and lists the catalog counts.

4. **Summarize Available Tools:**
   Once configured, report back to me confirming the setup is complete and list the tools now available in your context:
   - `search_skills`: Search agent skills by keyword, category, or tag.
   - `get_skill_doc`: Fetch complete markdown documentation and rules for any skill.
   - `search_mcp_servers`: Search catalog of MCP servers and capabilities.
   - `get_mcp_config`: Get configuration snippets for other MCP servers.
   - `calculate_overhead`: Estimate prompt token costs and context degradation risk.
   - `generate_agent_setup`: Generate client configuration blocks.

Please begin by asking for my Vercel domain if not already provided, or proceed immediately if you have access to it!
```
