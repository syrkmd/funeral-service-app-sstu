import { proxyApiClient } from "./client";

type RawRecord = Record<string, any>;

export type DashboardOverviewResponse = {
  uptime: string;
  totalRequests: number;
  currentRps: number;
  activeConnections: number;
  blockedRequests: number;
  upstreamStatus: string;
  averageLatencyMs: number;
};

export type MetricsResponse = {
  uptime: string;
  totalRequests: number;
  activeRequests: number;
  totalBlockedRequests: number;
  totalRateLimitedRequests: number;
  totalUpstreamErrors: number;
  requestsPerSecond: number;
  averageLatencyMs: number;
  bytesIn: number;
  bytesOut: number;
  activeConnections: number;
  requestsByStatus: Record<string, number>;
  requestsByMethod: Record<string, number>;
  requestsByPath: Record<string, number>;
  cacheHits: number;
  cacheMisses: number;
  cacheStores: number;
  cacheInvalidations: Record<string, number>;
};

export type ClientSnapshot = {
  ip: string;
  requests: number;
  bytesIn: number;
  bytesOut: number;
  blockedRequests: number;
  rateLimitedRequests: number;
};

export type DashboardClientsResponse = {
  topClientsByRequests: ClientSnapshot[];
  topClientsByTraffic: ClientSnapshot[];
  blockedClients: ClientSnapshot[];
  rateLimitedClients: ClientSnapshot[];
};

export type DashboardUpstreamResponse = {
  upstreamHealth: boolean;
  upstreamLatencyMs: number;
  upstreamErrors: number;
  lastUpstreamError?: string;
  totalUpstreamRequests: number;
  lastUpstreamStatus: number;
  lastUpdatedAt?: string;
};

export type RateLimitRule = {
  id: string;
  scope: "ip" | "subnet";
  value: string;
  rps: number;
  rpm: number;
  rph: number;
  rpd: number;
  cps: number;
  maxConnections: number;
  uploadBps: number;
  downloadBps: number;
  totalBytes: number;
  description?: string;
};

export type RateLimitBucketSnapshot = {
  key: string;
  scope: string;
  type: string;
  value: string;
  tokensRemaining: number;
  lastRefill: string;
};

export type BlockedClient = {
  ip: string;
  count: number;
};

export type DashboardRateLimitsResponse = {
  activeRateLimitRules: RateLimitRule[];
  currentBucketUsage: RateLimitBucketSnapshot[];
  violations: Record<string, number>;
  blockedIps: BlockedClient[];
};

export type DashboardIPRule = {
  id: string;
  type: "allowlist" | "denylist" | "graylist";
  value: string;
  description?: string;
};

export type DashboardIPAccessResponse = {
  allowlist: DashboardIPRule[];
  denylist: DashboardIPRule[];
  graylist: DashboardIPRule[];
  denyStatistics: Record<string, number>;
  matchedRulesStatistics: Record<string, number>;
};

export type ProxyDashboardMetrics = {
  totalRequests: number;
  errors: number;
  totalUpstreamErrors: number;
  activeClients: number;
  activeConnections: number;
  avgLatency: number;
  averageLatencyMs: number;
  traffic: number;
  requestsPerSecond: number;
  rps: number;
  bytesIn: number;
  bytesOut: number;
  cacheHits: number;
  cacheMisses: number;
  cacheStores: number;
  rateLimitViolations: number;
  blockedRequests: number;
  upstreamStatus: string;
};

function asRecord(value: unknown): RawRecord {
  return value && typeof value === "object" ? (value as RawRecord) : {};
}

function numberField(source: RawRecord, snake: string, camel: string = snake) {
  const value = source[snake] ?? source[camel];
  return typeof value === "number" ? value : Number(value || 0);
}

function stringField(source: RawRecord, snake: string, camel: string = snake) {
  const value = source[snake] ?? source[camel];
  return typeof value === "string" ? value : "";
}

function booleanField(source: RawRecord, snake: string, camel: string = snake) {
  return Boolean(source[snake] ?? source[camel]);
}

function mapField(source: RawRecord, snake: string, camel: string = snake): Record<string, number> {
  return asRecord(source[snake] ?? source[camel]);
}

function arrayField<T>(source: RawRecord, snake: string, camel: string, mapper: (item: RawRecord) => T): T[] {
  const value = source[snake] ?? source[camel];
  return Array.isArray(value) ? value.map((item) => mapper(asRecord(item))) : [];
}

function bytesToGB(bytes: number) {
  return Number((bytes / 1024 / 1024 / 1024).toFixed(2));
}

function sumValues(values: Record<string, number> | undefined) {
  return Object.values(values || {}).reduce((total, value) => total + Number(value || 0), 0);
}

function normalizeOverview(raw: unknown): DashboardOverviewResponse {
  const data = asRecord(raw);
  return {
    uptime: stringField(data, "uptime"),
    totalRequests: numberField(data, "total_requests", "totalRequests"),
    currentRps: numberField(data, "current_rps", "currentRps"),
    activeConnections: numberField(data, "active_connections", "activeConnections"),
    blockedRequests: numberField(data, "blocked_requests", "blockedRequests"),
    upstreamStatus: stringField(data, "upstream_status", "upstreamStatus"),
    averageLatencyMs: numberField(data, "average_latency_ms", "averageLatencyMs"),
  };
}

function normalizeMetrics(raw: unknown): MetricsResponse {
  const data = asRecord(raw);
  return {
    uptime: stringField(data, "uptime"),
    totalRequests: numberField(data, "total_requests", "totalRequests"),
    activeRequests: numberField(data, "active_requests", "activeRequests"),
    totalBlockedRequests: numberField(data, "total_blocked_requests", "totalBlockedRequests"),
    totalRateLimitedRequests: numberField(data, "total_rate_limited_requests", "totalRateLimitedRequests"),
    totalUpstreamErrors: numberField(data, "total_upstream_errors", "totalUpstreamErrors"),
    requestsPerSecond: numberField(data, "requests_per_second", "requestsPerSecond"),
    averageLatencyMs: numberField(data, "average_latency_ms", "averageLatencyMs"),
    bytesIn: numberField(data, "bytes_in", "bytesIn"),
    bytesOut: numberField(data, "bytes_out", "bytesOut"),
    activeConnections: numberField(data, "active_connections", "activeConnections"),
    requestsByStatus: mapField(data, "requests_by_status", "requestsByStatus"),
    requestsByMethod: mapField(data, "requests_by_method", "requestsByMethod"),
    requestsByPath: mapField(data, "requests_by_path", "requestsByPath"),
    cacheHits: numberField(data, "cache_hits", "cacheHits"),
    cacheMisses: numberField(data, "cache_misses", "cacheMisses"),
    cacheStores: numberField(data, "cache_stores", "cacheStores"),
    cacheInvalidations: mapField(data, "cache_invalidations", "cacheInvalidations"),
  };
}

function normalizeClient(raw: RawRecord): ClientSnapshot {
  return {
    ip: stringField(raw, "ip"),
    requests: numberField(raw, "requests"),
    bytesIn: numberField(raw, "bytes_in", "bytesIn"),
    bytesOut: numberField(raw, "bytes_out", "bytesOut"),
    blockedRequests: numberField(raw, "blocked_requests", "blockedRequests"),
    rateLimitedRequests: numberField(raw, "rate_limited_requests", "rateLimitedRequests"),
  };
}

function normalizeClients(raw: unknown): DashboardClientsResponse {
  const data = asRecord(raw);
  return {
    topClientsByRequests: arrayField(data, "top_clients_by_requests", "topClientsByRequests", normalizeClient),
    topClientsByTraffic: arrayField(data, "top_clients_by_traffic", "topClientsByTraffic", normalizeClient),
    blockedClients: arrayField(data, "blocked_clients", "blockedClients", normalizeClient),
    rateLimitedClients: arrayField(data, "rate_limited_clients", "rateLimitedClients", normalizeClient),
  };
}

function normalizeUpstream(raw: unknown): DashboardUpstreamResponse {
  const data = asRecord(raw);
  return {
    upstreamHealth: booleanField(data, "upstream_health", "upstreamHealth"),
    upstreamLatencyMs: numberField(data, "upstream_latency_ms", "upstreamLatencyMs"),
    upstreamErrors: numberField(data, "upstream_errors", "upstreamErrors"),
    lastUpstreamError: stringField(data, "last_upstream_error", "lastUpstreamError"),
    totalUpstreamRequests: numberField(data, "total_upstream_requests", "totalUpstreamRequests"),
    lastUpstreamStatus: numberField(data, "last_upstream_status", "lastUpstreamStatus"),
    lastUpdatedAt: stringField(data, "last_updated_at", "lastUpdatedAt"),
  };
}

function normalizeRateLimitRule(raw: RawRecord): RateLimitRule {
  return {
    id: stringField(raw, "id"),
    scope: (stringField(raw, "scope") || "ip") as RateLimitRule["scope"],
    value: stringField(raw, "value"),
    rps: numberField(raw, "rps"),
    rpm: numberField(raw, "rpm"),
    rph: numberField(raw, "rph"),
    rpd: numberField(raw, "rpd"),
    cps: numberField(raw, "cps"),
    maxConnections: numberField(raw, "max_connections", "maxConnections"),
    uploadBps: numberField(raw, "upload_bps", "uploadBps"),
    downloadBps: numberField(raw, "download_bps", "downloadBps"),
    totalBytes: numberField(raw, "total_bytes", "totalBytes"),
    description: stringField(raw, "description"),
  };
}

function normalizeBucket(raw: RawRecord): RateLimitBucketSnapshot {
  return {
    key: stringField(raw, "key"),
    scope: stringField(raw, "scope"),
    type: stringField(raw, "type"),
    value: stringField(raw, "value"),
    tokensRemaining: numberField(raw, "tokens_remaining", "tokensRemaining"),
    lastRefill: stringField(raw, "last_refill", "lastRefill"),
  };
}

function normalizeRateLimits(raw: unknown): DashboardRateLimitsResponse {
  const data = asRecord(raw);
  return {
    activeRateLimitRules: arrayField(data, "active_rate_limit_rules", "activeRateLimitRules", normalizeRateLimitRule),
    currentBucketUsage: arrayField(data, "current_bucket_usage", "currentBucketUsage", normalizeBucket),
    violations: mapField(data, "violations"),
    blockedIps: arrayField(data, "blocked_ips", "blockedIps", (item) => ({
      ip: stringField(item, "ip"),
      count: numberField(item, "count"),
    })),
  };
}

function normalizeIPRule(raw: RawRecord): DashboardIPRule {
  return {
    id: stringField(raw, "id"),
    type: (stringField(raw, "type") || "allowlist") as DashboardIPRule["type"],
    value: stringField(raw, "value"),
    description: stringField(raw, "description"),
  };
}

function normalizeIPAccess(raw: unknown): DashboardIPAccessResponse {
  const data = asRecord(raw);
  return {
    allowlist: arrayField(data, "allowlist", "allowlist", normalizeIPRule),
    denylist: arrayField(data, "denylist", "denylist", normalizeIPRule),
    graylist: arrayField(data, "graylist", "graylist", normalizeIPRule),
    denyStatistics: mapField(data, "deny_statistics", "denyStatistics"),
    matchedRulesStatistics: mapField(data, "matched_rules_statistics", "matchedRulesStatistics"),
  };
}

export function adaptDashboardMetrics(
  overview: DashboardOverviewResponse,
  metrics: MetricsResponse,
): ProxyDashboardMetrics {
  const totalRequests = overview.totalRequests || metrics.totalRequests;
  const activeConnections = overview.activeConnections || metrics.activeConnections;
  const averageLatencyMs = overview.averageLatencyMs || metrics.averageLatencyMs;
  const requestsPerSecond = overview.currentRps || metrics.requestsPerSecond;
  const bytesIn = metrics.bytesIn;
  const bytesOut = metrics.bytesOut;

  return {
    totalRequests,
    errors: metrics.totalUpstreamErrors,
    totalUpstreamErrors: metrics.totalUpstreamErrors,
    activeClients: activeConnections,
    activeConnections,
    avgLatency: Math.round(averageLatencyMs),
    averageLatencyMs,
    traffic: bytesToGB(bytesIn + bytesOut),
    requestsPerSecond,
    rps: requestsPerSecond,
    bytesIn,
    bytesOut,
    cacheHits: metrics.cacheHits,
    cacheMisses: metrics.cacheMisses,
    cacheStores: metrics.cacheStores,
    rateLimitViolations: metrics.totalRateLimitedRequests,
    blockedRequests: overview.blockedRequests || metrics.totalBlockedRequests,
    upstreamStatus: overview.upstreamStatus,
  };
}

export async function fetchMetrics(): Promise<ProxyDashboardMetrics> {
  const [overview, metrics] = await Promise.all([
    fetchDashboardOverview(),
    fetchRawMetrics(),
  ]);

  return adaptDashboardMetrics(overview, metrics);
}

export async function fetchRawMetrics(): Promise<MetricsResponse> {
  const response = await proxyApiClient.get("/api/metrics");
  return normalizeMetrics(response.data);
}

export async function fetchDashboardOverview(): Promise<DashboardOverviewResponse> {
  const response = await proxyApiClient.get("/api/dashboard/overview");
  return normalizeOverview(response.data);
}

export async function fetchDashboardClients(): Promise<DashboardClientsResponse> {
  const response = await proxyApiClient.get("/api/dashboard/clients");
  return normalizeClients(response.data);
}

export async function fetchDashboardUpstream(): Promise<DashboardUpstreamResponse> {
  const response = await proxyApiClient.get("/api/dashboard/upstream");
  return normalizeUpstream(response.data);
}

export async function fetchDashboardRateLimits(): Promise<DashboardRateLimitsResponse> {
  const response = await proxyApiClient.get("/api/dashboard/rate_limits");
  return normalizeRateLimits(response.data);
}

export async function fetchDashboardIPAccess(): Promise<DashboardIPAccessResponse> {
  const response = await proxyApiClient.get("/api/dashboard/ip_access");
  return normalizeIPAccess(response.data);
}

async function optional<T>(request: Promise<T>, empty: T): Promise<T> {
  try {
    return await request;
  } catch {
    return empty;
  }
}

export async function fetchDashboardSnapshot() {
  const [metrics, clients, upstream, rateLimits, ipAccess] = await Promise.all([
    fetchMetrics(),
    optional(fetchDashboardClients(), normalizeClients({})),
    optional(fetchDashboardUpstream(), normalizeUpstream({})),
    optional(fetchDashboardRateLimits(), normalizeRateLimits({})),
    optional(fetchDashboardIPAccess(), normalizeIPAccess({})),
  ]);

  return { metrics, clients, upstream, rateLimits, ipAccess };
}

export { sumValues };
