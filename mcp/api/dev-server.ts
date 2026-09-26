import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import mcpHandler from './mcp.js';
import healthHandler from './health.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;
const PUBLIC_DIR = path.resolve(__dirname, '../public');

const server = http.createServer(async (req, res) => {
  const url = new URL(req.url || '/', `http://${req.headers.host || 'localhost'}`);

  // 1. Health Endpoint
  if (url.pathname === '/api/health') {
    const webReq = new Request(url.toString(), {
      method: req.method,
      headers: req.headers as HeadersInit
    });
    const webRes = await healthHandler(webReq);
    res.writeHead(webRes.status, Object.fromEntries(webRes.headers.entries()));
    const body = await webRes.text();
    res.end(body);
    return;
  }

  // 2. MCP Endpoint (/api/mcp, /sse, /mcp)
  if (url.pathname === '/api/mcp' || url.pathname === '/sse' || url.pathname === '/mcp') {
    const chunks: Buffer[] = [];
    req.on('data', (c) => chunks.push(c));
    req.on('end', async () => {
      const rawBody = Buffer.concat(chunks);
      const webReq = new Request(url.toString(), {
        method: req.method,
        headers: req.headers as HeadersInit,
        body: req.method !== 'GET' && req.method !== 'HEAD' ? rawBody : undefined
      });

      try {
        const webRes = await mcpHandler(webReq);
        res.writeHead(webRes.status, Object.fromEntries(webRes.headers.entries()));
        if (webRes.body) {
          const reader = webRes.body.getReader();
          while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            res.write(value);
          }
        }
        res.end();
      } catch (err: any) {
        console.error('MCP Handler error:', err);
        res.writeHead(500, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: err.message }));
      }
    });
    return;
  }

  // 3. Static Files (public/)
  let filePath = path.join(PUBLIC_DIR, url.pathname === '/' ? 'index.html' : url.pathname);
  if (fs.existsSync(filePath) && fs.statSync(filePath).isFile()) {
    const ext = path.extname(filePath).toLowerCase();
    const mimeTypes: Record<string, string> = {
      '.html': 'text/html',
      '.css': 'text/css',
      '.js': 'text/javascript',
      '.json': 'application/json',
      '.svg': 'image/svg+xml',
      '.png': 'image/png'
    };
    res.writeHead(200, { 'Content-Type': mimeTypes[ext] || 'application/octet-stream' });
    fs.createReadStream(filePath).pipe(res);
    return;
  }

  // Fallback to index.html
  const fallback = path.join(PUBLIC_DIR, 'index.html');
  if (fs.existsSync(fallback)) {
    res.writeHead(200, { 'Content-Type': 'text/html' });
    fs.createReadStream(fallback).pipe(res);
    return;
  }

  res.writeHead(404, { 'Content-Type': 'text/plain' });
  res.end('Not Found');
});

server.listen(PORT, () => {
  console.log(`\n🚀 Skillyfiy MCP Server running locally at http://localhost:${PORT}`);
  console.log(`📊 Dashboard: http://localhost:${PORT}/`);
  console.log(`⚡ MCP Endpoint: http://localhost:${PORT}/api/mcp`);
  console.log(`💚 Health Check: http://localhost:${PORT}/api/health\n`);
});
