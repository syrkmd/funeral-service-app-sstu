import { useEffect } from "react";
import { fetchDashboardSnapshot, sumValues, type ClientSnapshot, type DashboardClientsResponse, type DashboardIPAccessResponse, type DashboardRateLimitsResponse, type DashboardUpstreamResponse } from "../../../api/metrics.api";
import { useMetricsStore } from "../../store/metricsStore";
import { useState } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

export function Dashboard() {
  const rpsData = useMetricsStore((state) => state.rpsData);
  const totalRequests = useMetricsStore((state) => state.totalRequests);
  const errors = useMetricsStore((state) => state.errors);
  const activeClients = useMetricsStore((state) => state.activeClients);
  const avgLatency = useMetricsStore((state) => state.avgLatency);
  const traffic = useMetricsStore((state) => state.traffic);
  const cacheHits = useMetricsStore((state) => state.cacheHits);
  const cacheMisses = useMetricsStore((state) => state.cacheMisses);
  const cacheStores = useMetricsStore((state) => state.cacheStores);
  const rateLimitViolations = useMetricsStore((state) => state.rateLimitViolations);
  const blockedRequests = useMetricsStore((state) => state.blockedRequests);
  const upstreamStatus = useMetricsStore((state) => state.upstreamStatus);

  const addRPSDataPoint = useMetricsStore((state) => state.addRPSDataPoint);
  const updateProxyMetrics = useMetricsStore((state) => state.updateProxyMetrics);
  const [clients, setClients] = useState<DashboardClientsResponse | null>(null);
  const [upstream, setUpstream] = useState<DashboardUpstreamResponse | null>(null);
  const [rateLimits, setRateLimits] = useState<DashboardRateLimitsResponse | null>(null);
  const [ipAccess, setIPAccess] = useState<DashboardIPAccessResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  // Real-time RPS updates
  useEffect(() => {
    const loadDashboard = async () => {
      try {
        const snapshot = await fetchDashboardSnapshot();
        updateProxyMetrics(snapshot.metrics);
        setClients(snapshot.clients);
        setUpstream(snapshot.upstream);
        setRateLimits(snapshot.rateLimits);
        setIPAccess(snapshot.ipAccess);
        setLoadError(null);
        addRPSDataPoint({
          time: new Date().toLocaleTimeString("ru-RU", {
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
          }),
          value: snapshot.metrics.rps,
        });
      } catch (error) {
        setLoadError(error instanceof Error ? error.message : "Failed to load dashboard");
      } finally {
        setIsLoading(false);
      }
    };

    loadDashboard();

    const interval = setInterval(loadDashboard, 1500);

    return () => clearInterval(interval);
  }, [addRPSDataPoint, updateProxyMetrics]);

  const ipRuleCount =
    (ipAccess?.allowlist || []).length +
    (ipAccess?.denylist || []).length +
    (ipAccess?.graylist || []).length;

  const stats = [
    { label: "Total Requests", value: totalRequests.toLocaleString(), change: "Live" },
    { label: "Upstream Errors", value: errors.toString(), change: upstreamStatus },
    { label: "Active Connections", value: activeClients.toLocaleString(), change: "Live" },
    { label: "Cache Hits / Misses", value: `${cacheHits} / ${cacheMisses}`, change: "Live" },
    { label: "Cache Stores", value: cacheStores.toString(), change: "Live" },
    { label: "Avg Latency", value: `${avgLatency}ms`, change: "Live" },
    { label: "Traffic", value: `${traffic}GB`, change: "Live" },
    { label: "Rate Limit Violations", value: rateLimitViolations.toString(), change: "Live" },
    { label: "Blocked Requests", value: blockedRequests.toString(), change: "Live" },
    { label: "IP Rules", value: ipRuleCount.toString(), change: "Live" },
  ];

  const chartData = rpsData.slice(-30);
  const topClients = getTopClients(clients);
  const ipStatsTotal = sumValues(ipAccess?.denyStatistics);
  const rateLimitViolationTotal = sumValues(rateLimits?.violations);

  return (
    <div className="space-y-8">
      {loadError && (
        <div className="bg-card border border-border rounded-lg p-4 text-sm text-destructive">
          {loadError}
        </div>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="bg-card border border-border rounded-lg p-6"
          >
            <div className="text-sm text-muted-foreground mb-2">
              {stat.label}
            </div>
            <div className="flex items-end justify-between">
              <div className="text-3xl text-foreground">{stat.value}</div>
              <div
                className={`text-sm ${
                  stat.change.startsWith("+")
                    ? "text-primary"
                    : stat.change.startsWith("-") &&
                      (stat.label.includes("Latency") ||
                        stat.label.includes("Error"))
                    ? "text-primary"
                    : "text-muted-foreground"
                }`}
              >
                {stat.change}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* RPS Chart */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Requests Per Second (RPS)</h3>
        {rpsData.length === 0 ? (
          <div className="h-64 flex items-center justify-center text-muted-foreground">
            {isLoading ? "Загрузка метрик..." : "Нет данных RPS"}
          </div>
        ) : (
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart
                data={chartData}
                margin={{ top: 10, right: 12, bottom: 0, left: -16 }}
              >
                <defs>
                  <linearGradient id="rpsGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.36} />
                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0.04} />
                  </linearGradient>
                </defs>
                <CartesianGrid stroke="hsl(var(--border))" strokeDasharray="3 3" vertical={false} />
                <XAxis
                  dataKey="time"
                  tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 11 }}
                  tickLine={false}
                  axisLine={{ stroke: "hsl(var(--border))" }}
                  minTickGap={18}
                />
                <YAxis
                  tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 11 }}
                  tickLine={false}
                  axisLine={false}
                  width={44}
                  domain={[
                    (dataMin: number) => Math.max(0, dataMin - Math.max(dataMin * 0.2, 0.2)),
                    (dataMax: number) => dataMax + Math.max(dataMax * 0.25, 0.5),
                  ]}
                  tickFormatter={(value) => Number(value).toFixed(1)}
                />
                <Tooltip
                  contentStyle={{
                    background: "hsl(var(--background))",
                    border: "1px solid hsl(var(--border))",
                    borderRadius: 8,
                    color: "hsl(var(--foreground))",
                  }}
                  labelStyle={{ color: "hsl(var(--muted-foreground))" }}
                  formatter={(value) => [`${Number(value).toFixed(2)} RPS`, "RPS"]}
                />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke="hsl(var(--primary))"
                  strokeWidth={2.5}
                  fill="url(#rpsGradient)"
                  fillOpacity={1}
                  dot={false}
                  activeDot={{ r: 4, fill: "hsl(var(--primary))", stroke: "hsl(var(--background))", strokeWidth: 2 }}
                  isAnimationActive
                  animationDuration={450}
                  animationEasing="ease-out"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
        <div className="flex justify-between mt-2 text-xs text-muted-foreground">
          <span>{rpsData[0]?.time || '00:00:00'}</span>
          <span>Live Updates</span>
          <span>{rpsData[rpsData.length - 1]?.time || '00:00:00'}</span>
        </div>
      </div>

      {/* Proxy Monitoring Sections */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-card border border-border rounded-lg p-6">
          <h3 className="text-lg mb-4 text-foreground">Upstream Status</h3>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <div className="text-muted-foreground mb-1">Health</div>
              <div className={upstream?.upstreamHealth ? "text-primary" : "text-destructive"}>
                {upstream?.upstreamHealth ? "Healthy" : "Degraded"}
              </div>
            </div>
            <div>
              <div className="text-muted-foreground mb-1">Latency</div>
              <div className="text-foreground">{Math.round(upstream?.upstreamLatencyMs || 0)}ms</div>
            </div>
            <div>
              <div className="text-muted-foreground mb-1">Errors</div>
              <div className="text-foreground">{upstream?.upstreamErrors || 0}</div>
            </div>
            <div>
              <div className="text-muted-foreground mb-1">Last Status</div>
              <div className="text-foreground">{upstream?.lastUpstreamStatus || "-"}</div>
            </div>
          </div>
        </div>

        <div className="bg-card border border-border rounded-lg p-6">
          <h3 className="text-lg mb-4 text-foreground">Security Events</h3>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <div className="text-muted-foreground mb-1">Rate Violations</div>
              <div className="text-foreground">{rateLimitViolationTotal.toLocaleString()}</div>
            </div>
            <div>
              <div className="text-muted-foreground mb-1">Blocked IPs</div>
              <div className="text-foreground">{(rateLimits?.blockedIps || []).length}</div>
            </div>
            <div>
              <div className="text-muted-foreground mb-1">IP Decisions</div>
              <div className="text-foreground">{ipStatsTotal.toLocaleString()}</div>
            </div>
            <div>
              <div className="text-muted-foreground mb-1">Active Rules</div>
              <div className="text-foreground">{(rateLimits?.activeRateLimitRules || []).length}</div>
            </div>
          </div>
        </div>
      </div>

      {/* Top Clients */}
      <div className="bg-card border border-border rounded-lg overflow-hidden">
        <div className="p-6 border-b border-border">
          <h3 className="text-lg text-foreground">Top Clients</h3>
        </div>
        <table className="w-full">
          <thead className="bg-secondary/50 border-b border-border">
            <tr>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">IP</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Requests</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Traffic</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Blocked</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Rate Limited</th>
            </tr>
          </thead>
          <tbody>
            {topClients.map((client) => (
              <tr key={client.ip} className="border-b border-border last:border-0 hover:bg-secondary/30 transition-colors">
                <td className="px-6 py-4 text-sm text-foreground font-mono">{client.ip}</td>
                <td className="px-6 py-4 text-sm text-foreground">{client.requests.toLocaleString()}</td>
                <td className="px-6 py-4 text-sm text-muted-foreground">
                  {((client.bytesIn + client.bytesOut) / 1024 / 1024).toFixed(2)} MB
                </td>
                <td className="px-6 py-4 text-sm text-muted-foreground">{client.blockedRequests}</td>
                <td className="px-6 py-4 text-sm text-muted-foreground">{client.rateLimitedRequests}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {topClients.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            Нет данных по клиентам
          </div>
        )}
      </div>

      {/* Recent Activity */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Rate Limit Violations</h3>
        {Object.keys(rateLimits?.violations || {}).length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            Нарушений не обнаружено
          </div>
        ) : (
          <div className="space-y-3">
            {Object.entries(rateLimits?.violations || {}).map(([type, count]) => (
              <div
                key={type}
                className="flex items-center justify-between py-3 border-b border-border last:border-0"
              >
                <div>
                  <div className="text-sm text-foreground">
                    {type}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Rate limiting event type
                  </div>
                </div>
                <div className="text-xs text-muted-foreground">
                  {count.toLocaleString()}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function getTopClients(clients: DashboardClientsResponse | null): ClientSnapshot[] {
  const source = [
    ...(clients?.topClientsByRequests || []),
    ...(clients?.topClientsByTraffic || []),
    ...(clients?.blockedClients || []),
    ...(clients?.rateLimitedClients || []),
  ];

  const byIP = new Map<string, ClientSnapshot>();
  source.forEach((client) => {
    const existing = byIP.get(client.ip);
    if (!existing || client.requests > existing.requests) {
      byIP.set(client.ip, client);
    }
  });

  return Array.from(byIP.values()).sort((a, b) => b.requests - a.requests).slice(0, 10);
}
