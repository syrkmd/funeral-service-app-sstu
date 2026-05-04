import { Link } from "react-router";
import { useOrdersStore, type OrderStatus } from "../../store/ordersStore";
import { useUserStore } from "../../store/userStore";

export function AccountOrders() {
  const currentUserPhone = useUserStore((state) => state.currentUserPhone);
  const allOrders = useOrdersStore((state) => state.orders);

  // Filter orders by current user's phone
  const orders = allOrders.filter((order) => order.phone === currentUserPhone);

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
      <span className={`px-3 py-1 rounded-full text-sm ${styles[status]}`}>
        {labels[status]}
      </span>
    );
  };

  if (orders.length === 0) {
    return (
      <div className="text-center py-16">
        <div className="mb-6">
          <svg
            className="w-20 h-20 text-muted-foreground mx-auto"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
        </div>
        <h2 className="text-xl text-foreground mb-2">Заказы не найдены</h2>
        <p className="text-muted-foreground mb-6">У вас пока нет оформленных заказов</p>
        <Link
          to="/order"
          className="inline-block bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:opacity-90 transition-opacity"
        >
          Оформить услугу
        </Link>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-3xl text-foreground mb-8">Мои заказы</h1>

      <div className="space-y-4">
        {orders.map((order) => (
          <div
            key={order.id}
            className="bg-card border border-border rounded-lg p-6 hover:shadow-sm transition-shadow"
          >
            <div className="flex items-start justify-between mb-4">
              <div>
                <h3 className="text-lg text-foreground mb-1">Заказ #{order.id}</h3>
                <div className="flex items-center gap-3">
                  <p className="text-sm text-muted-foreground">{order.date}</p>
                  {order.documents.length > 0 && (
                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                      <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
                        />
                      </svg>
                      {order.documents.length} {order.documents.length === 1 ? 'документ' : 'документа'}
                    </span>
                  )}
                </div>
              </div>
              <div className="flex flex-col items-end gap-2">
                {getStatusBadge(order.status)}
                {order.isPaid && (
                  <span className="text-xs text-primary">✓ Оплачен</span>
                )}
              </div>
            </div>

            <div className="flex items-center justify-between">
              <div className="text-2xl text-primary">{order.total.toLocaleString()} ₽</div>
              <Link
                to={`/account/orders/${order.id}`}
                className="px-4 py-2 bg-secondary text-secondary-foreground border border-border rounded-lg hover:bg-accent transition-colors"
              >
                Подробнее
              </Link>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
