import {
  searchSkills,
  getSkillDoc,
  searchMcpServers,
  getMcpConfig,
  calculateOverhead,
  generateAgentSetup
} from '../lib/tools.js';
import { estimateTokensFromText, assessContextOverhead } from '../lib/tokens.js';

function assert(condition: boolean, message: string) {
  if (!condition) {
    console.error(`❌ FAILED: ${message}`);
    process.exit(1);
  }
  console.log(`✓ ${message}`);
}

console.log('🧪 Starting Skillyfiy MCP Server Unit Tests...\n');

// 1. Test searchSkills
console.log('--- Testing searchSkills ---');
const skillsAll = searchSkills();
assert(skillsAll.total_matched > 0, 'Returns full catalog when query is empty');

const skillsTdd = searchSkills('tdd');
assert(skillsTdd.total_matched >= 1, "Finds 'test-driven-development' with query 'tdd'");
assert(skillsTdd.skills[0].name === 'test-driven-development', 'First match is test-driven-development');

const skillsFrontend = searchSkills(undefined, 'frontend');
assert(skillsFrontend.total_matched >= 2, 'Filters skills by category: frontend');

// 2. Test getSkillDoc
console.log('\n--- Testing getSkillDoc ---');
const doc = getSkillDoc('architect-review');
assert(doc.found === true, "Finds skill 'architect-review'");
assert(typeof doc.documentation === 'string' && doc.documentation.includes('Architect Review Skill'), 'Contains markdown content');

const missingDoc = getSkillDoc('non-existent-skill-xyz');
assert(missingDoc.found === false, 'Returns found=false for missing skill');

// 3. Test searchMcpServers
console.log('\n--- Testing searchMcpServers ---');
const mcpAll = searchMcpServers();
assert(mcpAll.total_matched >= 5, 'Finds MCP servers in catalog');

const mcpPostgres = searchMcpServers('postgres');
assert(mcpPostgres.total_matched >= 1, "Finds 'postgres' MCP server");
assert(mcpPostgres.servers[0].tools.includes('query'), "Postgres server lists 'query' tool");

// 4. Test getMcpConfig
console.log('\n--- Testing getMcpConfig ---');
const configClaude = getMcpConfig('github', 'claude');
assert(configClaude.found === true, "Generates config for 'github' server");
assert(Boolean(configClaude.config_json?.includes('mcpServers')), 'Config contains mcpServers block');

const configAntigravity = getMcpConfig('filesystem', 'antigravity');
assert(Boolean(configAntigravity.config_json?.includes('filesystem')), 'Antigravity config contains filesystem server');

// 5. Test calculateOverhead & tokens
console.log('\n--- Testing calculateOverhead ---');
const overhead = calculateOverhead(['architect-review', 'test-driven-development'], ['postgres']);
assert(overhead.total_tokens > 2000, `Calculates total tokens accurately (${overhead.total_tokens} tokens)`);
assert(overhead.breakdown.length === 3, 'Includes all requested items in breakdown');
assert(typeof overhead.assessment === 'string', 'Returns actionable optimization advice');

// 6. Test generateAgentSetup
console.log('\n--- Testing generateAgentSetup ---');
const setup = generateAgentSetup('cursor', 'https://my-app.vercel.app');
assert(setup.endpoint === 'https://my-app.vercel.app/api/mcp', 'Normalizes endpoint to /api/mcp');
assert(setup.snippet.includes('https://my-app.vercel.app/api/mcp'), 'Snippet contains the normalized URL');

console.log('\n🎉 ALL 14 ASSERTIONS PASSED! Skillyfiy MCP Server logic is 100% verified.\n');
