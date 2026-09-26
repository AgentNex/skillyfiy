import { SKILLS_CATALOG, AgentSkill } from './data/skills.js';
import { MCP_SERVERS_CATALOG, MCPServerInfo } from './data/mcp-servers.js';
import { assessContextOverhead, estimateTokensFromText } from './tokens.js';

export function searchSkills(query?: string, category?: string, limit = 10) {
  let results = [...SKILLS_CATALOG];

  if (category && category !== 'all') {
    const cat = category.toLowerCase().trim();
    results = results.filter((s) => s.category.toLowerCase() === cat);
  }

  if (query && query.trim() !== '') {
    const q = query.toLowerCase().trim();
    results = results.filter((s) => {
      return (
        s.name.toLowerCase().includes(q) ||
        s.description.toLowerCase().includes(q) ||
        s.tags.some((t) => t.toLowerCase().includes(q))
      );
    });
  }

  return {
    total_matched: results.length,
    skills: results.slice(0, limit).map((s) => ({
      name: s.name,
      description: s.description,
      category: s.category,
      risk: s.risk,
      tags: s.tags,
      estimated_tokens: s.tokens
    }))
  };
}

export function getSkillDoc(skillName: string) {
  const cleanName = skillName.toLowerCase().trim();
  const skill = SKILLS_CATALOG.find((s) => s.name.toLowerCase() === cleanName);

  if (!skill) {
    const available = SKILLS_CATALOG.map((s) => s.name).join(', ');
    return {
      found: false,
      error: `Skill '${skillName}' not found in registry.`,
      available_skills: available
    };
  }

  return {
    found: true,
    name: skill.name,
    category: skill.category,
    risk: skill.risk,
    tags: skill.tags,
    estimated_tokens: skill.tokens,
    documentation: skill.content
  };
}

export function searchMcpServers(query?: string, category?: string, limit = 10) {
  let results = [...MCP_SERVERS_CATALOG];

  if (category && category !== 'all') {
    const cat = category.toLowerCase().trim();
    results = results.filter((s) => s.category.toLowerCase() === cat);
  }

  if (query && query.trim() !== '') {
    const q = query.toLowerCase().trim();
    results = results.filter((s) => {
      return (
        s.name.toLowerCase().includes(q) ||
        s.description.toLowerCase().includes(q) ||
        s.tools.some((t) => t.toLowerCase().includes(q))
      );
    });
  }

  return {
    total_matched: results.length,
    servers: results.slice(0, limit).map((s) => ({
      name: s.name,
      description: s.description,
      category: s.category,
      transport: s.transport,
      tools_count: s.tools.length,
      tools: s.tools,
      estimated_tokens: s.tokens
    }))
  };
}

export function getMcpConfig(serverName: string, clientType = 'claude') {
  const cleanName = serverName.toLowerCase().trim();
  const server = MCP_SERVERS_CATALOG.find((s) => s.name.toLowerCase() === cleanName);

  if (!server) {
    return {
      found: false,
      error: `MCP Server '${serverName}' not found in catalog.`
    };
  }

  let configSnippet: any;

  switch (clientType.toLowerCase()) {
    case 'cursor':
    case 'windsurf':
    case 'claude':
    default:
      configSnippet = {
        mcpServers: {
          [server.name]: {
            command: server.command,
            args: server.args,
            ...(server.env ? { env: server.env } : {})
          }
        }
      };
      break;
    case 'antigravity':
    case 'gemini':
      configSnippet = {
        [server.name]: {
          command: server.command,
          args: server.args,
          ...(server.env ? { env: server.env } : {})
        }
      };
      break;
  }

  return {
    found: true,
    server: server.name,
    client: clientType,
    tools: server.tools,
    estimated_overhead: `~${server.tokens} prompt tokens`,
    config_json: JSON.stringify(configSnippet, null, 2)
  };
}

export function calculateOverhead(skillNames: string[] = [], mcpNames: string[] = []) {
  let totalTokens = 0;
  const breakdown: Array<{ type: 'SKILL' | 'MCP'; name: string; tokens: number }> = [];

  for (const name of skillNames) {
    const s = SKILLS_CATALOG.find((x) => x.name.toLowerCase() === name.toLowerCase().trim());
    const t = s ? s.tokens : 1200; // fallback conservative token estimate
    breakdown.push({ type: 'SKILL', name, tokens: t });
    totalTokens += t;
  }

  for (const name of mcpNames) {
    const m = MCP_SERVERS_CATALOG.find((x) => x.name.toLowerCase() === name.toLowerCase().trim());
    const t = m ? m.tokens : 500;
    breakdown.push({ type: 'MCP', name, tokens: t });
    totalTokens += t;
  }

  const assessment = assessContextOverhead(totalTokens);

  return {
    total_tokens: totalTokens,
    total_items: breakdown.length,
    impact_level: assessment.level,
    context_percentage: `${assessment.impactPercentage}% of 32k window`,
    assessment: assessment.advice,
    breakdown
  };
}

export function generateAgentSetup(clientType: string, endpointUrl: string) {
  const url = endpointUrl.trim().replace(/\/+$/, '');
  const sseEndpoint = url.endsWith('/api/mcp') ? url : `${url}/api/mcp`;

  const client = clientType.toLowerCase();

  let targetFile = '';
  let configBlock: any = {};

  if (client.includes('claude')) {
    targetFile = 'claude_desktop_config.json (macOS: ~/Library/Application Support/Claude/ | Windows: %APPDATA%/Claude/)';
    configBlock = {
      mcpServers: {
        skillyfiy: {
          url: sseEndpoint
        }
      }
    };
  } else if (client.includes('cursor')) {
    targetFile = '.cursor/mcp.json or Global Cursor Settings';
    configBlock = {
      mcpServers: {
        skillyfiy: {
          url: sseEndpoint
        }
      }
    };
  } else if (client.includes('windsurf')) {
    targetFile = '~/.codeium/windsurf/mcp_config.json';
    configBlock = {
      mcpServers: {
        skillyfiy: {
          serverUrl: sseEndpoint
        }
      }
    };
  } else {
    targetFile = 'mcp_config.json';
    configBlock = {
      mcpServers: {
        skillyfiy: {
          url: sseEndpoint
        }
      }
    };
  }

  return {
    client: clientType,
    endpoint: sseEndpoint,
    configuration_file_location: targetFile,
    snippet: JSON.stringify(configBlock, null, 2),
    instructions: `Add the above JSON snippet to your ${clientType} configuration file and restart the application.`
  };
}
