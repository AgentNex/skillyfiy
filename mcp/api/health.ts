import { SKILLS_CATALOG } from '../lib/data/skills.js';
import { MCP_SERVERS_CATALOG } from '../lib/data/mcp-servers.js';

export default function handler(req: Request) {
  const healthData = {
    status: 'healthy',
    server: 'skillyfiy-mcp',
    version: '1.0.0',
    timestamp: new Date().toISOString(),
    catalog: {
      skills_indexed: SKILLS_CATALOG.length,
      mcp_servers_indexed: MCP_SERVERS_CATALOG.length
    },
    endpoints: {
      mcp_sse: '/api/mcp',
      mcp_http: '/api/mcp',
      dashboard: '/'
    }
  };

  return new Response(JSON.stringify(healthData, null, 2), {
    status: 200,
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*'
    }
  });
}

export const GET = handler;
