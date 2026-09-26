import { createMcpHandler } from 'mcp-handler';
import { z } from 'zod';
import {
  searchSkills,
  getSkillDoc,
  searchMcpServers,
  getMcpConfig,
  calculateOverhead,
  generateAgentSetup
} from '../lib/tools.js';

const handler = createMcpHandler(
  (server) => {
    // 1. Search Skills Tool
    server.registerTool(
      'search_skills',
      {
        description: 'Search agent skills by keyword, category, or tag with estimated token overhead.',
        inputSchema: {
          query: z.string().optional().describe('Keyword to search in skill name, description, or tags'),
          category: z.string().optional().describe('Filter by category: architecture, frontend, testing, backend, devops, ai-agents, database'),
          limit: z.number().optional().describe('Max results to return (default: 10)')
        }
      },
      async ({ query, category, limit }) => {
        const data = searchSkills(query, category, limit);
        return {
          content: [{ type: 'text', text: JSON.stringify(data, null, 2) }]
        };
      }
    );

    // 2. Get Skill Documentation Tool
    server.registerTool(
      'get_skill_doc',
      {
        description: 'Fetch the complete documentation, usage guidelines, prompts, and rules of any agent skill.',
        inputSchema: {
          skill_name: z.string().describe('Exact name of the skill (e.g. architect-review, test-driven-development, animejs-animation)')
        }
      },
      async ({ skill_name }) => {
        const data = getSkillDoc(skill_name);
        return {
          content: [{ type: 'text', text: JSON.stringify(data, null, 2) }]
        };
      }
    );

    // 3. Search MCP Servers Tool
    server.registerTool(
      'search_mcp_servers',
      {
        description: 'Search catalog of MCP servers by name, description, tools, or category.',
        inputSchema: {
          query: z.string().optional().describe('Search keyword for server name, tools, or capabilities'),
          category: z.string().optional().describe('Filter by category: database, devtools, cloud, search, productivity, memory'),
          limit: z.number().optional().describe('Max results to return (default: 10)')
        }
      },
      async ({ query, category, limit }) => {
        const data = searchMcpServers(query, category, limit);
        return {
          content: [{ type: 'text', text: JSON.stringify(data, null, 2) }]
        };
      }
    );

    // 4. Get MCP Server Configuration Tool
    server.registerTool(
      'get_mcp_config',
      {
        description: 'Get the copy-paste configuration block for any MCP server for Claude, Cursor, Antigravity, or Windsurf.',
        inputSchema: {
          server_name: z.string().describe('Name of the MCP server (e.g. filesystem, github, postgres, cloudflare, chrome-devtools)'),
          client_type: z.string().optional().describe('Target client format: claude, cursor, antigravity, windsurf')
        }
      },
      async ({ server_name, client_type }) => {
        const data = getMcpConfig(server_name, client_type);
        return {
          content: [{ type: 'text', text: JSON.stringify(data, null, 2) }]
        };
      }
    );

    // 5. Calculate Token Overhead Tool
    server.registerTool(
      'calculate_overhead',
      {
        description: 'Calculate prompt token overhead and context degradation risk for active skills and MCP servers.',
        inputSchema: {
          skills: z.array(z.string()).optional().describe('List of active skill names'),
          mcp_servers: z.array(z.string()).optional().describe('List of active MCP server names')
        }
      },
      async ({ skills, mcp_servers }) => {
        const data = calculateOverhead(skills, mcp_servers);
        return {
          content: [{ type: 'text', text: JSON.stringify(data, null, 2) }]
        };
      }
    );

    // 6. Generate Agent Setup Snippet Tool
    server.registerTool(
      'generate_agent_setup',
      {
        description: 'Generate ready-to-use MCP configuration file snippet to connect this Skillyfiy MCP Server to any AI client.',
        inputSchema: {
          client_type: z.string().describe('Client type: claude, cursor, windsurf, antigravity, cline'),
          endpoint_url: z.string().describe('Your deployed Vercel endpoint URL (e.g. https://your-skillyfiy.vercel.app)')
        }
      },
      async ({ client_type, endpoint_url }) => {
        const data = generateAgentSetup(client_type, endpoint_url);
        return {
          content: [{ type: 'text', text: JSON.stringify(data, null, 2) }]
        };
      }
    );
  },
  {
    serverInfo: {
      name: 'skillyfiy-mcp',
      version: '1.0.0'
    }
  }
);

export const GET = handler;
export const POST = handler;
export default handler;
