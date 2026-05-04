import { useState } from "react";
import { Link } from "react-router";
import { useOrdersStore, type OrderStatus } from "../../store/ordersStore";
import { useOrderPolling } from "../../hooks/useOrderPolling";
import { useOrderNotifications } from "../../hooks/useOrderNotifications";

export function Orders() {
  // Enable real-time polling for orders
  useOrderPolling();

  // Show notifications when new orders arrive
  useOrderNotifications();

  const orders = useOrdersStore((state) => state.orders);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const filteredOrders = orders.filter((order) => {
    const matchesSearch =
      order.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      order.client.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      order.client.phone.includes(searchQuery);

    const matchesStatus = statusFilter === "all" || order.status === statusFilter;

    return matchesSearch && matchesStatus;
  });

  const getStatusBadge = (status: OrderStatus) => {
    const styles = {
      processing: "bg-secondary text-secondary-foreground border border-border",
      confirmed: "bg-primary/10 text-primary border border-primary/30",
      completed: "bg-primary/20 text-primary border border-primary/40",
      cancelled: "bg-destructive/10 text-destructive border border-destructive/30",
    };

    const labels = {
      processing: "В обработке",
      confirmed: "Подтверждён",
      completed: "Завершён",
      cancelled: "Отменён",
    };

    return (
      <span className={`px-3 py-1 rounded-full text-xs ${styles[status]}`}>
        {labels[status]}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      {/* Filters */}
      <div className="bg-card border border-border rounded-lg p-6">
        <div className="flex flex-col md:flex-row gap-4">
          <input
            type="text"
            placeholder="Поиск по ID, имени или телефону..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="flex-1 px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
          />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="all">Все статусы</option>
            <option value="processing">В обработке</option>
            <option value="confirmed">Подтверждён</option>
            <option value="completed">Завершён</option>
            <option value="cancelled">Отменён</option>
          </select>
        </div>
      </div>

      {/* Orders Table */}
      <div className="bg-card border border-border rounded-lg overflow-hidden">
        <table className="w-full">
          <thead className="bg-secondary/50 border-b border-border">
            <tr>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">ID заказа</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Клиент</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Дата</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Сумма</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Статус</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Оплата</th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">Действия</th>
            </tr>
          </thead>
          <tbody>
            {filteredOrders.map((order) => (
              <tr key={order.id} className="border-b border-border last:border-0 hover:bg-secondary/30 transition-colors">
                <td className="px-6 py-4 text-sm text-foreground font-mono">{order.id}</td>
                <td className="px-6 py-4">
                  <div className="text-sm text-foreground">{order.client.name}</div>
                  <div className="text-xs text-muted-foreground">{order.client.phone}</div>
                </td>
                <td className="px-6 py-4 text-sm text-muted-foreground">{order.date}</td>
                <td className="px-6 py-4 text-sm text-primary">{order.total.toLocaleString()} ₽</td>
                <td className="px-6 py-4">{getStatusBadge(order.status)}</td>
                <td className="px-6 py-4">
                  <span
                    className={`px-3 py-1 rounded-full text-xs ${
                      order.isPaid
                        ? "bg-primary/20 text-primary border border-primary/40"
                        : "bg-secondary text-secondary-foreground border border-border"
                    }`}
                  >
                    {order.isPaid ? "Оплачен" : "Не оплачен"}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <Link
                    to={`/admin/orders/${order.id}`}
                    className="text-sm text-primary hover:underline"
                  >
                    Подробнее
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {filteredOrders.length === 0 && (
          <div className="text-center py-12 text-muted-foreground">
            Заказы не найдены
          </div>
        )}
      </div>
    </div>
  );
}
