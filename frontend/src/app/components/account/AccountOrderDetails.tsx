import { useState, type FormEvent } from "react";
import { useParams, useNavigate } from "react-router";
import { CreditCard, LoaderCircle, LockKeyhole } from "lucide-react";
import { useOrdersStore } from "../../store/ordersStore";
import { useUserStore } from "../../store/userStore";

export function AccountOrderDetails() {
  const { orderId } = useParams();
  const navigate = useNavigate();
  const currentUserPhone = useUserStore((state) => state.currentUserPhone);
  const payOrder = useOrdersStore((state) => state.payOrder);
  const [cardNumber, setCardNumber] = useState("");
  const [expiryDate, setExpiryDate] = useState("");
  const [cvv, setCvv] = useState("");
  const [isPaying, setIsPaying] = useState(false);
  const [paymentError, setPaymentError] = useState<string | null>(null);
  const [paymentSuccess, setPaymentSuccess] = useState(false);

  const order = useOrdersStore((state) =>
    state.orders.find((o) => o.id === orderId)
  );

  const handleDownload = (documentName: string) => {
    alert(`Скачивание: ${documentName}`);
  };

  const handleCardNumberChange = (value: string) => {
    setCardNumber(value.replace(/\D/g, "").slice(0, 16));
  };

  const handleExpiryChange = (value: string) => {
    const digits = value.replace(/\D/g, "").slice(0, 4);
    setExpiryDate(digits.length > 2 ? `${digits.slice(0, 2)}/${digits.slice(2)}` : digits);
  };

  const handlePayment = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!orderId) return;

    setIsPaying(true);
    setPaymentError(null);
    setPaymentSuccess(false);

    try {
      await payOrder(orderId, { cardNumber, cvv, expiryDate });
      setPaymentSuccess(true);
      setCardNumber("");
      setExpiryDate("");
      setCvv("");
    } catch (error) {
      setPaymentError(
        error instanceof Error ? error.message : "Не удалось выполнить оплату",
      );
    } finally {
      setIsPaying(false);
    }
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

      {/* Ceremony Info */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Детали церемонии</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Дата и время:</span>
            <p className="text-foreground">
              {order.serviceDate || "Не указано"}
              {order.serviceTime ? ` в ${order.serviceTime}` : ""}
            </p>
          </div>
          <div>
            <span className="text-muted-foreground">Место проведения:</span>
            <p className="text-foreground">{order.serviceAddress || "Не указано"}</p>
          </div>
          <div>
            <span className="text-muted-foreground">Место захоронения:</span>
            <p className="text-foreground">{order.cemeteryPlotCode || "Не выбрано"}</p>
          </div>
          <div>
            <span className="text-muted-foreground">Примечания:</span>
            <p className="text-foreground">{order.cemeteryNotes || "Нет примечаний"}</p>
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

      {!order.isPaid && order.status !== "cancelled" && (
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="flex items-start gap-3 mb-5">
            <div className="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center flex-shrink-0">
              <CreditCard className="w-5 h-5 text-primary" />
            </div>
            <div>
              <h3 className="text-lg text-foreground">Оплата банковской картой</h3>
              <p className="text-sm text-muted-foreground">
                К оплате: {order.total.toLocaleString()} ₽
              </p>
            </div>
          </div>

          <form onSubmit={handlePayment} className="space-y-4 max-w-xl">
            <div>
              <label htmlFor="card-number" className="block text-sm text-foreground mb-2">
                Номер карты
              </label>
              <input
                id="card-number"
                value={cardNumber}
                onChange={(event) => handleCardNumberChange(event.target.value)}
                inputMode="numeric"
                autoComplete="cc-number"
                placeholder="0000 0000 0000 0000"
                required
                pattern="\d{16}"
                className="w-full h-11 px-3 bg-background border border-border rounded-md text-foreground outline-none focus:border-primary"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="expiry-date" className="block text-sm text-foreground mb-2">
                  Срок действия
                </label>
                <input
                  id="expiry-date"
                  value={expiryDate}
                  onChange={(event) => handleExpiryChange(event.target.value)}
                  inputMode="numeric"
                  autoComplete="cc-exp"
                  placeholder="ММ/ГГ"
                  required
                  pattern="(0[1-9]|1[0-2])/\d{2}"
                  className="w-full h-11 px-3 bg-background border border-border rounded-md text-foreground outline-none focus:border-primary"
                />
              </div>
              <div>
                <label htmlFor="card-cvv" className="block text-sm text-foreground mb-2">
                  CVV
                </label>
                <input
                  id="card-cvv"
                  type="password"
                  value={cvv}
                  onChange={(event) => setCvv(event.target.value.replace(/\D/g, "").slice(0, 3))}
                  inputMode="numeric"
                  autoComplete="cc-csc"
                  placeholder="000"
                  required
                  pattern="\d{3}"
                  className="w-full h-11 px-3 bg-background border border-border rounded-md text-foreground outline-none focus:border-primary"
                />
              </div>
            </div>

            {paymentError && (
              <p role="alert" className="text-sm text-destructive">
                {paymentError}
              </p>
            )}

            <button
              type="submit"
              disabled={isPaying}
              className="h-11 px-5 bg-primary text-primary-foreground rounded-md hover:opacity-90 disabled:opacity-60 transition-opacity inline-flex items-center justify-center gap-2"
            >
              {isPaying ? (
                <LoaderCircle className="w-4 h-4 animate-spin" />
              ) : (
                <LockKeyhole className="w-4 h-4" />
              )}
              {isPaying ? "Оплата..." : `Оплатить ${order.total.toLocaleString()} ₽`}
            </button>
          </form>
        </div>
      )}

      {(order.isPaid || paymentSuccess) && (
        <div className="border border-primary/30 bg-primary/10 rounded-lg p-5">
          <p className="text-primary">Оплата успешно выполнена</p>
          <p className="text-sm text-muted-foreground mt-1">
            Заказ отмечен как оплаченный.
          </p>
        </div>
      )}

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
