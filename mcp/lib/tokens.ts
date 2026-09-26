/**
 * Token overhead calculation ported from Skillyfiy's Go engine.
 * Average tokenization ratio for code and mixed markdown prompt context:
 * ~3.25 bytes per token (or ~4 characters per token).
 */
export const BYTES_PER_TOKEN = 3.25;

export function estimateTokensFromText(text: string): number {
  if (!text) return 0;
  const byteLength = new TextEncoder().encode(text).length;
  return Math.ceil(byteLength / BYTES_PER_TOKEN);
}

export function estimateTokensFromBytes(bytes: number): number {
  if (bytes <= 0) return 0;
  return Math.ceil(bytes / BYTES_PER_TOKEN);
}

export interface ContextImpact {
  totalTokens: number;
  level: 'optimal' | 'moderate' | 'high' | 'critical';
  impactPercentage: number; // based on typical 32k or 128k active attention window
  statusColor: string;
  advice: string;
}

export function assessContextOverhead(totalTokens: number, contextLimit = 32768): ContextImpact {
  const percentage = Math.min(100, Math.round((totalTokens / contextLimit) * 100));

  if (totalTokens < 4000) {
    return {
      totalTokens,
      level: 'optimal',
      impactPercentage: percentage,
      statusColor: '#10B981', // Mint Green
      advice: 'Context overhead is minimal. AI agent reasoning and needle-in-a-haystack recall remain sharp.'
    };
  }

  if (totalTokens < 12000) {
    return {
      totalTokens,
      level: 'moderate',
      impactPercentage: percentage,
      statusColor: '#06B6D4', // Cyan
      advice: 'Noticeable prompt overhead. Ensure inactive skills are pruned using Skillyfiy to preserve speed.'
    };
  }

  if (totalTokens < 24000) {
    return {
      totalTokens,
      level: 'high',
      impactPercentage: percentage,
      statusColor: '#F59E0B', // Amber
      advice: 'High context consumption. Risk of reasoning degradation, instruction dilution, and higher API token costs.'
    };
  }

  return {
    totalTokens,
    level: 'critical',
    impactPercentage: percentage,
    statusColor: '#EF4444', // Red
    advice: 'Critical context bloat! Immediate pruning recommended. Purge unused MCP tools and redundant skill guides.'
  };
}
