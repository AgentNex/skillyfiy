export interface MCPServerInfo {
  name: string;
  description: string;
  category: 'database' | 'devtools' | 'cloud' | 'search' | 'productivity' | 'memory';
  transport: 'stdio' | 'sse' | 'both';
  tools: string[];
  command: string;
  args: string[];
  env?: Record<string, string>;
  npmPackage?: string;
  tokens: number;
}

export const MCP_SERVERS_CATALOG: MCPServerInfo[] = [
  {
    name: 'filesystem',
    description: 'Official Model Context Protocol server providing secure, sandboxed local file reading, editing, directory inspection, and searching.',
    category: 'devtools',
    transport: 'stdio',
    tools: ['read_file', 'read_multiple_files', 'write_file', 'create_directory', 'list_directory', 'move_file', 'search_files', 'get_file_info'],
    command: 'npx',
    args: ['-y', '@modelcontextprotocol/server-filesystem', '/path/to/allowed/directory'],
    npmPackage: '@modelcontextprotocol/server-filesystem',
    tokens: 420
  },
  {
    name: 'github',
    description: 'Official GitHub MCP server allowing agents to browse code, create issues, search repositories, manage pull requests, and trigger Actions.',
    category: 'devtools',
    transport: 'stdio',
    tools: ['create_issue', 'get_issue', 'update_issue', 'create_pull_request', 'fork_repository', 'create_branch', 'push_files', 'search_repositories'],
    command: 'npx',
    args: ['-y', '@modelcontextprotocol/server-github'],
    env: {
      GITHUB_PERSONAL_ACCESS_TOKEN: '<YOUR_GITHUB_TOKEN>'
    },
    npmPackage: '@modelcontextprotocol/server-github',
    tokens: 580
  },
  {
    name: 'postgres',
    description: 'Read-only and transactional PostgreSQL MCP server enabling direct SQL schema inspection, EXPLAIN planning, and query execution.',
    category: 'database',
    transport: 'stdio',
    tools: ['query', 'list_tables', 'describe_table'],
    command: 'npx',
    args: ['-y', '@modelcontextprotocol/server-postgres', 'postgresql://user:password@localhost:5432/dbname'],
    npmPackage: '@modelcontextprotocol/server-postgres',
    tokens: 380
  },
  {
    name: 'cloudflare',
    description: 'Comprehensive Cloudflare server managing Workers deployments, KV namespaces, D1 SQL databases, and R2 object storage buckets.',
    category: 'cloud',
    transport: 'stdio',
    tools: ['workers_list', 'workers_get_worker', 'd1_database_query', 'r2_bucket_create', 'kv_namespace_get'],
    command: 'npx',
    args: ['-y', '@cloudflare/mcp-server-cloudflare'],
    env: {
      CLOUDFLARE_API_TOKEN: '<YOUR_API_TOKEN>',
      CLOUDFLARE_ACCOUNT_ID: '<YOUR_ACCOUNT_ID>'
    },
    npmPackage: '@cloudflare/mcp-server-cloudflare',
    tokens: 650
  },
  {
    name: 'chrome-devtools',
    description: 'Automate Chrome via DevTools Protocol: navigate, evaluate JavaScript, take screenshots, inspect console errors, and simulate mobile viewports.',
    category: 'devtools',
    transport: 'stdio',
    tools: ['navigate_page', 'take_screenshot', 'evaluate_script', 'click', 'type_text', 'list_console_messages', 'lighthouse_audit'],
    command: 'npx',
    args: ['-y', '@modelcontextprotocol/server-puppeteer'],
    npmPackage: '@modelcontextprotocol/server-puppeteer',
    tokens: 720
  },
  {
    name: 'sequential-thinking',
    description: 'Step-by-step dynamic reasoning server enabling agents to break down multi-step problems, revise hypotheses, and verify intermediate logic.',
    category: 'memory',
    transport: 'stdio',
    tools: ['sequentialthinking'],
    command: 'npx',
    args: ['-y', '@modelcontextprotocol/server-sequential-thinking'],
    npmPackage: '@modelcontextprotocol/server-sequential-thinking',
    tokens: 280
  },
  {
    name: 'engram',
    description: 'Cross-session persistent agent memory system supporting full-text observation search, topic consolidation, and compaction recovery.',
    category: 'memory',
    transport: 'stdio',
    tools: ['mem_save', 'mem_search', 'mem_context', 'mem_session_start', 'mem_session_end', 'mem_update'],
    command: 'engram-mcp',
    args: [],
    tokens: 490
  },
  {
    name: 'exa',
    description: 'Neural, embeddings-based web search and clean webpage extraction designed specifically for AI agent research and source verification.',
    category: 'search',
    transport: 'stdio',
    tools: ['web_search_exa', 'web_fetch_exa'],
    command: 'npx',
    args: ['-y', 'exa-mcp-server'],
    env: {
      EXA_API_KEY: '<YOUR_EXA_KEY>'
    },
    npmPackage: 'exa-mcp-server',
    tokens: 310
  }
];
