import { useEffect } from "react";
import { fetchRPSData } from "../../../api/metrics.api";
import { useMetricsStore } from "../../store/metricsStore";
import { useOrdersStore } from "../../store/ordersStore";
import { useOrderPolling } from "../../hooks/useOrderPolling";
import { buildActivityLog, type ActivityLogEntry } from "../../utils/activityLog";

export function Dashboard() {
  // Enable real-time polling for orders
  useOrderPolling();

  const rpsData = useMetricsStore((state) => state.rpsData);
  const totalRequests = useMetricsStore((state) => state.totalRequests);
  const errors = useMetricsStore((state) => state.errors);
  const activeClients = useMetricsStore((state) => state.activeClients);
  const ordersToday = useMetricsStore((state) => state.ordersToday);
  const avgLatency = useMetricsStore((state) => state.avgLatency);
  const traffic = useMetricsStore((state) => state.traffic);
  const orders = useOrdersStore((state) => state.orders);
  const activityLog = buildActivityLog(orders);

  const addRPSDataPoint = useMetricsStore((state) => state.addRPSDataPoint);
  const incrementTotalRequests = useMetricsStore((state) => state.incrementTotalRequests);

  // Real-time RPS updates
  useEffect(() => {
    const interval = setInterval(async () => {
      const dataPoint = await fetchRPSData();
      addRPSDataPoint(dataPoint);
      incrementTotalRequests(dataPoint.value);
    }, 3000); // Update every 3 seconds

    return () => clearInterval(interval);
  }, [addRPSDataPoint, incrementTotalRequests]);

  const stats = [
    { label: "Total Requests", value: totalRequests.toLocaleString(), change: "+12.5%" },
    { label: "Errors", value: errors.toString(), change: "-5.2%" },
    { label: "Active Clients", value: activeClients.toLocaleString(), change: "+8.1%" },
    { label: "Orders Today", value: ordersToday.toString(), change: `+${ordersToday}` },
    { label: "Avg Latency", value: `${avgLatency}ms`, change: "-8%" },
    { label: "Traffic", value: `${traffic}GB`, change: "+15%" },
  ];

  const getActivityLabel = (type: ActivityLogEntry['type']) => {
    switch (type) {
      case 'ORDER_CREATED':
        return 'Новый заказ';
      case 'STATUS_CHANGED':
        return 'Статус изменён';
      case 'PAYMENT_CONFIRMED':
        return 'Оплата подтверждена';
      default:
        return 'Событие';
    }
  };

  const getRelativeTime = (timestamp: number) => {
    const diff = Date.now() - timestamp;
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);

    if (minutes < 1) return 'только что';
    if (minutes < 60) return `${minutes} мин назад`;
    if (hours < 24) return `${hours} час${hours === 1 ? '' : hours < 5 ? 'а' : 'ов'} назад`;
    return new Date(timestamp).toLocaleDateString('ru-RU');
  };

  // Calculate max value for chart scaling
  const maxValue = Math.max(...rpsData.map(d => d.value), 100);

  return (
    <div className="space-y-8">
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
        <div className="h-64 flex items-end justify-between gap-2">
          {rpsData.map((dataPoint, i) => (
            <div key={i} className="flex-1 flex flex-col items-center gap-1">
              <div
                className="w-full bg-primary/20 rounded-t transition-all duration-300"
                style={{ height: `${(dataPoint.value / maxValue) * 100}%` }}
              />
              <div className="text-xs text-primary text-center">
                {dataPoint.value}
              </div>
            </div>
          ))}
        </div>
        <div className="flex justify-between mt-2 text-xs text-muted-foreground">
          <span>{rpsData[0]?.time || '00:00:00'}</span>
          <span>Live Updates</span>
          <span>{rpsData[rpsData.length - 1]?.time || '00:00:00'}</span>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Последняя активность</h3>
        {activityLog.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            Нет активности
          </div>
        ) : (
          <div className="space-y-3">
            {activityLog.map((activity) => (
              <div
                key={activity.id}
                className="flex items-center justify-between py-3 border-b border-border last:border-0"
              >
                <div>
                  <div className="text-sm text-foreground">
                    {getActivityLabel(activity.type)}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {activity.user}
                    {activity.orderId && ` • Заказ ${activity.orderId}`}
                  </div>
                </div>
                <div className="text-xs text-muted-foreground">
                  {getRelativeTime(activity.timestamp)}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
