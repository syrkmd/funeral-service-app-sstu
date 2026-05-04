import { useState } from "react";

type LogEntry = {
  id: number;
  ip: string;
  time: string;
  url: string;
  status: number;
  reason: string;
  level: "INFO" | "ERROR";
};

const mockLogs: LogEntry[] = [
  {
    id: 1,
    ip: "192.168.1.1",
    time: "2026-05-04 14:30:45",
    url: "/api/orders",
    status: 200,
    reason: "Success",
    level: "INFO",
  },
  {
    id: 2,
    ip: "10.0.0.5",
    time: "2026-05-04 14:29:12",
    url: "/api/auth",
    status: 401,
    reason: "Unauthorized",
    level: "ERROR",
  },
  {
    id: 3,
    ip: "172.16.0.10",
    time: "2026-05-04 14:28:33",
    url: "/api/orders/123",
    status: 200,
    reason: "Success",
    level: "INFO",
  },
  {
    id: 4,
    ip: "192.168.1.50",
    time: "2026-05-04 14:27:05",
    url: "/api/upload",
    status: 500,
    reason: "Internal Server Error",
    level: "ERROR",
  },
  {
    id: 5,
    ip: "10.0.0.8",
    time: "2026-05-04 13:25:18",
    url: "/api/orders",
    status: 200,
    reason: "Success",
    level: "INFO",
  },
  {
    id: 6,
    ip: "192.168.1.1",
    time: "2026-05-04 10:24:42",
    url: "/api/services",
    status: 200,
    reason: "Success",
    level: "INFO",
  },
  {
    id: 7,
    ip: "10.0.0.5",
    time: "2026-05-03 18:15:22",
    url: "/api/orders",
    status: 403,
    reason: "Blocked",
    level: "ERROR",
  },
];

export function Logs() {
  const [levelFilter, setLevelFilter] = useState<string>("all");
  const [ipFilter, setIpFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [timeFilter, setTimeFilter] = useState<string>("all");

  const filteredLogs = mockLogs.filter((log) => {
    // Level filter
    if (levelFilter !== "all" && log.level !== levelFilter) return false;

    // IP filter
    if (ipFilter && !log.ip.includes(ipFilter)) return false;

    // Status filter
    if (statusFilter === "success" && (log.status < 200 || log.status >= 300))
      return false;
    if (statusFilter === "blocked" && log.status < 400) return false;

    // Time filter
    const logDate = new Date(log.time);
    const now = new Date();
    if (timeFilter === "hour") {
      const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
      if (logDate < oneHourAgo) return false;
    }
    if (timeFilter === "day") {
      const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);
      if (logDate < oneDayAgo) return false;
    }

    return true;
  });

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return "text-primary";
    if (status >= 400) return "text-destructive";
    return "text-muted-foreground";
  };

  const getLevelBadge = (level: LogEntry["level"]) => {
    const styles = {
      INFO: "bg-primary/20 text-primary border border-primary/40",
      ERROR: "bg-destructive/20 text-destructive border border-destructive/40",
    };

    return (
      <span className={`px-3 py-1 rounded-full text-xs ${styles[level]}`}>
        {level}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      {/* Filters */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Фильтры</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm text-foreground mb-2">
              Уровень
            </label>
            <select
              value={levelFilter}
              onChange={(e) => setLevelFilter(e.target.value)}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="all">Все уровни</option>
              <option value="INFO">INFO</option>
              <option value="ERROR">ERROR</option>
            </select>
          </div>

          <div>
            <label className="block text-sm text-foreground mb-2">IP адрес</label>
            <input
              type="text"
              placeholder="192.168.1.1"
              value={ipFilter}
              onChange={(e) => setIpFilter(e.target.value)}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>

          <div>
            <label className="block text-sm text-foreground mb-2">Статус</label>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="all">Все</option>
              <option value="success">Успешные (2xx)</option>
              <option value="blocked">Заблокированные (4xx+)</option>
            </select>
          </div>

          <div>
            <label className="block text-sm text-foreground mb-2">Период</label>
            <select
              value={timeFilter}
              onChange={(e) => setTimeFilter(e.target.value)}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="all">Все время</option>
              <option value="hour">Последний час</option>
              <option value="day">Последний день</option>
            </select>
          </div>
        </div>

        {(levelFilter !== "all" ||
          ipFilter ||
          statusFilter !== "all" ||
          timeFilter !== "all") && (
          <div className="mt-4">
            <button
              onClick={() => {
                setLevelFilter("all");
                setIpFilter("");
                setStatusFilter("all");
                setTimeFilter("all");
              }}
              className="text-sm text-primary hover:underline"
            >
              Сбросить фильтры
            </button>
          </div>
        )}
      </div>

      {/* Logs Table */}
      <div className="bg-card border border-border rounded-lg overflow-hidden">
        <div className="p-4 bg-secondary/30 border-b border-border">
          <p className="text-sm text-muted-foreground">
            Найдено записей: {filteredLogs.length} из {mockLogs.length}
          </p>
        </div>
        <table className="w-full">
          <thead className="bg-secondary/50 border-b border-border">
            <tr>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                Level
              </th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                IP
              </th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                Time
              </th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                URL
              </th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                Reason
              </th>
            </tr>
          </thead>
          <tbody>
            {filteredLogs.map((log) => (
              <tr
                key={log.id}
                className="border-b border-border last:border-0 hover:bg-secondary/30 transition-colors"
              >
                <td className="px-6 py-4">{getLevelBadge(log.level)}</td>
                <td className="px-6 py-4 text-sm text-foreground font-mono">
                  {log.ip}
                </td>
                <td className="px-6 py-4 text-sm text-muted-foreground">
                  {log.time}
                </td>
                <td className="px-6 py-4 text-sm text-foreground font-mono">
                  {log.url}
                </td>
                <td className="px-6 py-4 text-sm">
                  <span className={getStatusColor(log.status)}>{log.status}</span>
                </td>
                <td className="px-6 py-4 text-sm text-muted-foreground">
                  {log.reason}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {filteredLogs.length === 0 && (
          <div className="text-center py-12 text-muted-foreground">
            Логи не найдены
          </div>
        )}
      </div>
    </div>
  );
}
