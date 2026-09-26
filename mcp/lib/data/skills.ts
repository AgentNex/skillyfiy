export interface AgentSkill {
  name: string;
  description: string;
  category: 'architecture' | 'frontend' | 'testing' | 'backend' | 'devops' | 'ai-agents' | 'database';
  risk: 'safe' | 'low' | 'medium';
  tags: string[];
  tokens: number;
  content: string;
}

export const SKILLS_CATALOG: AgentSkill[] = [
  {
    name: 'architect-review',
    description: 'Master software architect specializing in scalable microservices, clean architecture, domain modeling, and technical debt audit.',
    category: 'architecture',
    risk: 'safe',
    tags: ['architecture', 'system-design', 'ddd', 'clean-code'],
    tokens: 1420,
    content: `# Architect Review Skill

Comprehensive architectural review framework evaluating decoupling, scalability, data flow boundaries, and operational complexity.

## Principles
1. **Separation of Concerns**: Domain logic must remain independent of framework and transport layers.
2. **Failure Isolation**: High-frequency failure paths must have circuit breakers and explicit recovery mechanisms.
3. **Data Boundaries**: Eventual consistency vs strict ACID guarantees should be aligned with business domain constraints.

## Review Checklist
- Check for leaky abstractions across repository and service layers.
- Verify that shared libraries do not introduce hidden circular dependencies.
- Audit synchronization overhead in high-throughput pipelines.`
  },
  {
    name: 'test-driven-development',
    description: 'Strict Red-Green-Refactor test-driven development workflow to ensure bug-free, regression-resistant implementations.',
    category: 'testing',
    risk: 'safe',
    tags: ['tdd', 'unit-tests', 'refactor', 'quality'],
    tokens: 1250,
    content: `# Test-Driven Development (TDD) Skill

Execute disciplined test-first development across any programming language.

## Workflow Phases
1. **Red**: Write a minimal failing test expressing the desired behavior. Verify that it fails for the expected reason.
2. **Green**: Write the simplest possible implementation that satisfies the test suite. No premature optimization.
3. **Refactor**: Clean up the implementation, eliminate duplication, and preserve all green assertions.

## Invariant Rules
- Never write implementation code without a failing test case demonstrating its necessity.
- Treat test code with the same architectural standards as production code.`
  },
  {
    name: 'animejs-animation',
    description: 'High-performance JavaScript animation skill using Anime.js for smooth SVG morphing, staggered layouts, and orchestrated timelines.',
    category: 'frontend',
    risk: 'safe',
    tags: ['animation', 'javascript', 'svg', 'frontend'],
    tokens: 1100,
    content: `# Anime.js Animation Skill

Patterns for creating fluid, award-winning web animations using Anime.js.

## Core Capabilities
- Staggered grid animations with directional delay calculations.
- Seamless SVG path morphing with responsive canvas fallbacks.
- Hardware-accelerated CSS transform orchestrations.

## Performance Guardrails
- Always animate \`transform\` (translate, scale, rotate) and \`opacity\` to stay on the GPU compositor thread.
- Avoid animating layout-triggering properties (\`width\`, \`height\`, \`top\`, \`margin\`).`
  },
  {
    name: 'framer-motion',
    description: 'Production patterns for React motion design: layout animations, shared layout transitions, AnimatePresence, and gestures.',
    category: 'frontend',
    risk: 'safe',
    tags: ['react', 'animation', 'framer-motion', 'ui'],
    tokens: 1350,
    content: `# Framer Motion Skill

Interactive gesture handling, exit transitions, and auto-layout transitions for modern React applications.

## Key Patterns
- **Shared Layout Transitions**: Animate elements between different component boundaries using \`layoutId\`.
- **AnimatePresence**: Cleanly animate mounting and unmounting DOM nodes.
- **Spring Physics**: Tune stiffness and damping to prevent jarring oscillations on mobile screens.`
  },
  {
    name: 'cloudflare-workers-expert',
    description: 'Cloudflare Workers and Serverless Edge architecture: KV, R2, D1, Durable Objects, and sub-millisecond edge API routing.',
    category: 'devops',
    risk: 'safe',
    tags: ['cloudflare', 'workers', 'edge', 'serverless', 'hono'],
    tokens: 1680,
    content: `# Cloudflare Workers Expert Skill

Build and deploy globally distributed edge services with instant cold starts and minimal latency.

## Edge Principles
- Utilize \`Hono\` or Web Standard \`fetch\` event handlers.
- Offload asset storage to R2 buckets with pre-signed upload URLs.
- Manage consistent distributed state using Durable Objects with WebSocket pub/sub.
- Protect endpoints with Cloudflare Turnstile bot detection.`
  },
  {
    name: 'agent-memory',
    description: 'Hybrid persistent memory architecture for AI agents: short-term buffers, session recall, vector embeddings, and topic distillation.',
    category: 'ai-agents',
    risk: 'safe',
    tags: ['memory', 'agents', 'persistence', 'context-management'],
    tokens: 1540,
    content: `# Agent Memory Skill

Manage long-running agent state across context compactions and multiple conversation sessions.

## Architecture
- **Passive Memory Capture**: Extract decisions, bug fixes, and user constraints into persistent storage.
- **Compaction Recovery**: Restore active session goals immediately upon context resets.
- **Topic Key Deduplication**: Group evolving architectures under stable topic keys instead of fragmenting memories.`
  },
  {
    name: 'database-admin',
    description: 'PostgreSQL, SQLite, and MySQL database administration: indexing strategies, query plan EXPLAIN analysis, and zero-downtime migrations.',
    category: 'database',
    risk: 'safe',
    tags: ['database', 'postgresql', 'sql', 'indexes', 'performance'],
    tokens: 1480,
    content: `# Database Admin Skill

Ensure database reliability, high transaction throughput, and data integrity under heavy loads.

## Rules of Thumb
- Always use multi-column indexes matching the query's \`WHERE\` and \`ORDER BY\` cardinality.
- Run \`EXPLAIN (ANALYZE, BUFFERS)\` to identify sequential disk scans.
- Wrap table schema migrations in transactional blocks with safe lock timeouts.`
  },
  {
    name: 'insforge',
    description: 'Full BaaS integration with InsForge SDK: PostgreSQL CRUD, OAuth/JWT auth, edge functions, storage, and real-time WebSockets.',
    category: 'backend',
    risk: 'safe',
    tags: ['insforge', 'baas', 'database', 'auth', 'storage'],
    tokens: 1620,
    content: `# InsForge SDK Integration Skill

Patterns for building full-stack web and mobile apps with the InsForge BaaS platform.

## Crucial SDK Patterns
- Database inserts take an array: \`client.database.from('table').insert([{ ... }])\`.
- For storage uploads, persist both the returned \`url\` and \`key\`.
- Real-time subscriptions listen to PostgreSQL CDC events via WebSocket pub/sub.`
  },
  {
    name: 'security-auditor',
    description: 'Comprehensive cybersecurity audit skill: OWASP Top 10, sanitization, JWT verification, CORS policy, and zeroize sensitive keys.',
    category: 'backend',
    risk: 'safe',
    tags: ['security', 'audit', 'owasp', 'jwt', 'auth'],
    tokens: 1590,
    content: `# Security Auditor Skill

Proactive security scanning for APIs, authentication handlers, and web applications.

## Verification Checklist
1. **Input Sanitization**: Ensure parameterized SQL queries and schema validation using Zod.
2. **CORS Configuration**: Restrict allowed origins, disallowing wildcards on credentialed routes.
3. **Secret Zeroization**: Scrub passwords and tokens from memory immediately after verification.`
  },
  {
    name: 'browser-harness',
    description: 'Headless browser automation via Chrome DevTools Protocol (CDP) for visual verification, interactive testing, and deep scraping.',
    category: 'testing',
    risk: 'low',
    tags: ['browser', 'cdp', 'testing', 'scraping', 'devtools'],
    tokens: 1390,
    content: `# Browser Harness Skill

Direct interaction with running browser instances through CDP or Playwright for verified visual state audits.

## Interaction Flow
- Navigate to authenticated endpoints without clearing session cookies.
- Capture deterministic full-page screenshots and DOM snapshots.
- Evaluate client console logs to detect unhandled exceptions and network 4xx/5xx errors.`
  }
];
