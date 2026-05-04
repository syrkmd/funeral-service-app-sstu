import { useParams, useNavigate } from "react-router";
import { useOrdersStore } from "../../store/ordersStore";
import { useUserStore } from "../../store/userStore";

export function AccountOrderDetails() {
  const { orderId } = useParams();
  const navigate = useNavigate();
  const currentUserPhone = useUserStore((state) => state.currentUserPhone);

  const order = useOrdersStore((state) =>
    state.orders.find((o) => o.id === orderId)
  );

  const handleDownload = (documentName: string) => {
    alert(`Скачивание: ${documentName}`);
  };

  // Security check: verify order belongs to current user
  if (!order || order.phone !== currentUserPhone) {
    return (
      <div className="text-center py-16">
        <h2 className="text-xl text-foreground mb-2">Заказ не найден</h2>
        <p className="text-muted-foreground mb-6">
          Заказ с номером {orderId} не существует или вам не принадлежит
        </p>
        <button
          onClick={() => navigate("/account")}
          className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:opacity-90 transition-opacity"
        >
          Вернуться к заказам
        </button>
      </div>
    );
  }

  const getStatusBadge = () => {
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
      <span className={`px-3 py-1 rounded-full text-sm ${styles[order.status]}`}>
        {labels[order.status]}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <button
          onClick={() => navigate('/account')}
          className="text-sm text-muted-foreground hover:text-foreground mb-4"
        >
          ← Назад к заказам
        </button>
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-3xl text-foreground mb-2">Заказ #{order.id}</h1>
            <p className="text-sm text-muted-foreground">{order.date}</p>
          </div>
          <div className="flex flex-col items-end gap-2">
            {getStatusBadge()}
            <span
              className={`px-3 py-1 rounded-full text-sm ${
                order.isPaid
                  ? "bg-primary/20 text-primary border border-primary/40"
                  : "bg-secondary text-secondary-foreground border border-border"
              }`}
            >
              {order.isPaid ? "Оплачен" : "Не оплачен"}
            </span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Client Info */}
        <div className="bg-card border border-border rounded-lg p-6">
          <h3 className="text-lg mb-4 text-foreground">Данные клиента</h3>
          <div className="space-y-3 text-sm">
            <div>
              <span className="text-muted-foreground">ФИО:</span>
              <p className="text-foreground">{order.client.name}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Телефон:</span>
              <p className="text-foreground">{order.client.phone}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Email:</span>
              <p className="text-foreground">{order.client.email}</p>
            </div>
          </div>
        </div>

        {/* Deceased Info */}
        <div className="bg-card border border-border rounded-lg p-6">
          <h3 className="text-lg mb-4 text-foreground">Данные умершего</h3>
          <div className="space-y-3 text-sm">
            <div>
              <span className="text-muted-foreground">ФИО:</span>
              <p className="text-foreground">{order.deceased.name}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Дата рождения:</span>
              <p className="text-foreground">{order.deceased.dateOfBirth}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Дата смерти:</span>
              <p className="text-foreground">{order.deceased.dateOfDeath}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Services */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Услуги</h3>
        <div className="space-y-2">
          {order.services.map((service, index) => (
            <div key={index} className="flex justify-between py-2 border-b border-border last:border-0">
              <span className="text-foreground">{service.name}</span>
              <span className="text-primary">{service.price.toLocaleString()} ₽</span>
            </div>
          ))}
        </div>
      </div>

      {/* Products */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Товары</h3>
        <div className="space-y-2">
          {order.products.map((product, index) => (
            <div key={index} className="flex justify-between py-2 border-b border-border last:border-0">
              <span className="text-foreground">{product.name}</span>
              <span className="text-primary">{product.price.toLocaleString()} ₽</span>
            </div>
          ))}
        </div>
      </div>

      {/* Total */}
      <div className="bg-secondary/50 border border-border rounded-lg p-6">
        <div className="flex justify-between items-center">
          <span className="text-xl text-foreground">Итоговая сумма</span>
          <span className="text-3xl text-primary">{order.total.toLocaleString()} ₽</span>
        </div>
      </div>

      {/* Documents */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Документы</h3>

        {order.documents.length === 0 ? (
          <p className="text-muted-foreground text-sm">Документы пока не добавлены</p>
        ) : (
          <div className="space-y-4">
            {order.documents.map((document) => (
              <div
                key={document.id}
                className="flex items-start justify-between p-4 bg-background border border-border rounded-lg"
              >
                <div className="flex items-start gap-4">
                  <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg
                      className="w-6 h-6 text-primary"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
                      />
                    </svg>
                  </div>
                  <div>
                    <h4 className="text-foreground mb-1">{document.name}</h4>
                    <div className="flex items-center gap-3 text-sm text-muted-foreground">
                      <span>{document.type}</span>
                      <span>•</span>
                      <span>{document.size}</span>
                      <span>•</span>
                      <span>{document.date}</span>
                    </div>
                  </div>
                </div>

                <button
                  onClick={() => handleDownload(document.name)}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                    />
                  </svg>
                  Скачать
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
