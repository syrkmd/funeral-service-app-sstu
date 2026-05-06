import { useState } from "react";
import { useParams, useNavigate } from "react-router";
import { useOrdersStore, type OrderStatus } from "../../store/ordersStore";

const predefinedFiles = [
  { value: "death_certificate.pdf", label: "Свидетельство о смерти" },
  { value: "contract.pdf", label: "Договор на оказание услуг" },
  { value: "payment_receipt.pdf", label: "Квитанция об оплате" },
  { value: "medical_report.pdf", label: "Медицинское заключение" },
];

export function OrderDetails() {
  const { orderId } = useParams();
  const navigate = useNavigate();
  const [showAddDocument, setShowAddDocument] = useState(false);
  const [newDocName, setNewDocName] = useState("");
  const [newDocType, setNewDocType] = useState("PDF");
  const [newDocFile, setNewDocFile] = useState("");
  const [statusUpdateMessage, setStatusUpdateMessage] = useState(false);

  const order = useOrdersStore((state) =>
    state.orders.find((o) => o.id === orderId)
  );
  const updateOrderStatus = useOrdersStore((state) => state.updateOrderStatus);
  const updateOrderPayment = useOrdersStore((state) => state.updateOrderPayment);
  const addDocument = useOrdersStore((state) => state.addDocument);
  const removeDocument = useOrdersStore((state) => state.removeDocument);

  const handleStatusChange = async (newStatus: OrderStatus) => {
    if (orderId && order) {
      await updateOrderStatus(orderId, newStatus);
      setStatusUpdateMessage(true);
      setTimeout(() => setStatusUpdateMessage(false), 3000);
    }
  };

  const handlePaymentChange = async () => {
    if (orderId && order) {
      const newPaidStatus = !order.isPaid;
      await updateOrderPayment(orderId, newPaidStatus);
    }
  };

  const handleAddDocument = async () => {
    if (!orderId || !newDocName.trim() || !newDocFile) return;

    const newDocument = {
      id: Date.now(),
      name: newDocName.trim(),
      type: newDocType,
      size: "1.2 MB", // Mock size
      date: new Date().toISOString().split("T")[0],
    };

    await addDocument(orderId, newDocument);

    // Reset form
    setNewDocName("");
    setNewDocType("PDF");
    setNewDocFile("");
    setShowAddDocument(false);
  };

  const handleRemoveDocument = async (documentId: number) => {
    if (!orderId) return;
    if (confirm("Удалить этот документ?")) {
      await removeDocument(orderId, documentId);
    }
  };

  if (!order) {
    return (
      <div className="text-center py-16">
        <h2 className="text-xl text-foreground mb-2">Заказ не найден</h2>
        <p className="text-muted-foreground mb-6">
          Заказ с номером {orderId} не существует
        </p>
        <button
          onClick={() => navigate("/admin/orders")}
          className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:opacity-90 transition-opacity"
        >
          Вернуться к заказам
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <button
            onClick={() => navigate('/admin/orders')}
            className="text-sm text-muted-foreground hover:text-foreground mb-2"
          >
            ← Назад к заказам
          </button>
          <h2 className="text-2xl text-foreground">Заказ {order.id}</h2>
          <p className="text-sm text-muted-foreground">{order.date}</p>
        </div>
        <div className="flex gap-3 items-center">
          <div className="relative">
            <select
              value={order.status}
              onChange={(e) => handleStatusChange(e.target.value as OrderStatus)}
              className="px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="processing">В обработке</option>
              <option value="confirmed">Подтверждён</option>
              <option value="completed">Завершён</option>
              <option value="cancelled">Отменён</option>
            </select>
            {statusUpdateMessage && (
              <div className="absolute top-full mt-2 left-0 whitespace-nowrap bg-primary text-primary-foreground px-3 py-1 rounded text-sm">
                Статус обновлён
              </div>
            )}
          </div>
          <button
            onClick={handlePaymentChange}
            disabled={order.isPaid}
            className={`px-4 py-2 rounded-lg transition-colors ${
              order.isPaid
                ? "bg-primary/20 text-primary border border-primary/40 cursor-default"
                : "bg-secondary text-secondary-foreground border border-border hover:bg-accent"
            }`}
          >
            {order.isPaid ? "Оплачен" : "Отметить оплаченным"}
          </button>
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
          <span className="text-xl text-foreground">Итого</span>
          <span className="text-3xl text-primary">{order.total.toLocaleString()} ₽</span>
        </div>
      </div>

      {/* Documents */}
      <div className="bg-card border border-border rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg text-foreground">Документы</h3>
          <button
            onClick={() => setShowAddDocument(!showAddDocument)}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-sm"
          >
            {showAddDocument ? "Отмена" : "Добавить документ"}
          </button>
        </div>

        {/* Add Document Form */}
        {showAddDocument && (
          <div className="mb-6 p-4 bg-secondary/30 border border-border rounded-lg">
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-foreground mb-2">
                  Название документа *
                </label>
                <input
                  type="text"
                  value={newDocName}
                  onChange={(e) => setNewDocName(e.target.value)}
                  className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="Например: Свидетельство о смерти"
                />
              </div>
              <div>
                <label className="block text-sm text-foreground mb-2">Тип</label>
                <input
                  type="text"
                  value={newDocType}
                  onChange={(e) => setNewDocType(e.target.value)}
                  className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="PDF"
                />
              </div>
              <div>
                <label className="block text-sm text-foreground mb-2">
                  Файл *
                </label>
                <select
                  value={newDocFile}
                  onChange={(e) => setNewDocFile(e.target.value)}
                  className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value="">Выберите файл</option>
                  {predefinedFiles.map((file) => (
                    <option key={file.value} value={file.value}>
                      {file.label}
                    </option>
                  ))}
                </select>
              </div>
              <button
                onClick={handleAddDocument}
                disabled={!newDocName.trim() || !newDocFile}
                className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Добавить
              </button>
            </div>
          </div>
        )}

        {/* Documents List */}
        {order.documents.length === 0 ? (
          <p className="text-muted-foreground text-sm">Документы не добавлены</p>
        ) : (
          <div className="space-y-3">
            {order.documents.map((document) => (
              <div
                key={document.id}
                className="flex items-center justify-between p-4 bg-background border border-border rounded-lg"
              >
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg
                      className="w-5 h-5 text-primary"
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
                    <h4 className="text-foreground text-sm">{document.name}</h4>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1">
                      <span>{document.type}</span>
                      <span>•</span>
                      <span>{document.size}</span>
                      <span>•</span>
                      <span>{document.date}</span>
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => handleRemoveDocument(document.id)}
                  className="px-4 py-2 text-sm text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                >
                  Удалить
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
